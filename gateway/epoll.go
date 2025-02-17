package gateway

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/lxh-3260/plato/common/config"
	"golang.org/x/sys/unix"
)

// 全局对象
var ep *ePool    // epoll池，epoll轮询器池(epoll数量和cpu核数对应)
var tcpNum int32 // 当前服务允许接入的最大tcp连接数

// epoll 对象池
type ePool struct {
	// eid    int // epoll id
	eChan  chan *connection
	tables sync.Map                         // 并发安全的map，与java的concurrentHashMap类似
	eSize  int                              // epollsize，epoll池的大小
	done   chan struct{}                    // 用于关闭epoll，资源回收
	ln     *net.TCPListener                 // TCP监听器
	f      func(c *connection, ep *epoller) // 回调runProc函数
}

// epoller 对象 轮询器(轮询器数量和cpu核数对应)，每个轮询器有自己的fd和连接表
type epoller struct {
	fd            int
	fdToConnTable sync.Map
}

func initEpoll(ln *net.TCPListener, f func(c *connection, ep *epoller)) {
	setLimit() // 修改fd最大数量上限
	ep = newEPool(ln, f)
	ep.createAcceptProcess()
	ep.startEPool()
}

func newEPool(ln *net.TCPListener, cb func(c *connection, ep *epoller)) *ePool {
	return &ePool{
		eChan:  make(chan *connection, config.GetGatewayEpollerChanNum()), // 100
		done:   make(chan struct{}),
		eSize:  config.GetGatewayEpollerNum(), // 等于cpu核数
		tables: sync.Map{},
		ln:     ln,
		f:      cb,
	}
}

// 生产者
// 创建一个专门处理 accept 事件的协程，与当前cpu的核数对应，能够发挥最大功效
// 多个goroutine监听Accept channel，但是只有一个goroutine会成功消费accept，其他的goroutine会阻塞在Accept方法上
func (e *ePool) createAcceptProcess() {
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(i int) {
			for {
				// TODO: 给某个epoll对象加序号，用于区分epoll对象
				conn, e := e.ln.AcceptTCP() // 接受一个新的来自客户端的tcp连接
				// 限流熔断，当超过某个TCP连接数时，直接关闭连接
				if !checkTcp() {
					_ = conn.Close()
					continue
				}
				setTcpConifg(conn) // 开启keepalive
				if e != nil {
					if ne, ok := e.(net.Error); ok && (ne.Timeout() || ne.Temporary()) {
						log.Printf("accept err: %v", e)
						continue
					}
					log.Printf("Epoll %d: accept err: %v", i, e)
				}
				c := NewConnection(conn) // go无法直接获取fd，需要通过反射获取
				ep.addTask(c)            // 向channel中添加一个新的连接task，等待epoll池中的某个核的epoll处理
			}
		}(i)
	}
}

func (e *ePool) startEPool() {
	for i := 0; i < e.eSize; i++ {
		go e.startEProc()
	}
}

// 消费者
// 轮询器池 处理器
func (e *ePool) startEProc() {
	ep, err := newEpoller()
	if err != nil {
		panic(err)
	}
	// 监听连接创建事件
	go func() {
		for {
			select {
			case <-e.done: // 资源回收，优雅地关闭epoller
				return
			case conn := <-e.eChan: // epoller池中某个核的epoll接受到新的连接task，放入eChan中，在这里被消费
				if err := ep.add(conn); err != nil {
					fmt.Printf("failed to add connection %v\n", err)
					conn.Close() //登录未成功直接关闭连接
					continue
				}
				addTcpNum()
				fmt.Printf("tcpNum:%d\n", tcpNum)
				fmt.Printf("EpollerPool new connection[%v] tcpSize:%d\n", conn.RemoteAddr(), tcpNum)
			}
		}
	}()
	// wait的逻辑，轮询器在这里轮询等待, 当有wait发生时则调用回调函数去处理
	for {
		select {
		case <-e.done:
			return
		default:
			connections, err := ep.wait(200) // 200ms 一次轮询避免 忙轮询，返回值是200ms内: epoll事件fd对应的的connections切片
			if err != nil && err != syscall.EINTR {
				fmt.Printf("failed to epoll wait %v\n", err)
				continue
			}
			for _, conn := range connections {
				if conn == nil {
					break
				}
				e.f(conn, ep)
			}
		}
	}
}

func (e *ePool) addTask(c *connection) {
	e.eChan <- c
}

func newEpoller() (*epoller, error) {
	fd, err := unix.EpollCreate1(0) // epoll对象的fd，不是connection的fd
	if err != nil {
		return nil, err
	}
	return &epoller{
		fd: fd,
	}, nil
}

// TODO: 默认水平触发模式,可采用非阻塞FD,优化边沿触发模式
func (e *epoller) add(conn *connection) error {
	// Extract file descriptor associated with the connection
	fd := conn.fd
	// 将文件描述符添加到 epoll 实例中进行监控
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{ // 区分epoll对象的fd和connection的fd
		Events: unix.EPOLLIN | unix.EPOLLHUP, // 监听连接有消息到达的读写消息读事件、连接挂起（等待io）事件
		Fd:     int32(fd),                    // e.fd是epoller的fd，fd是connection的fd
	})
	if err != nil {
		return err
	}
	e.fdToConnTable.Store(fd, conn) // 存储conn.fd和connection的映射关系，因为wait事件触发时，epoll只能知道fd，需要通过fd找到对应的connection，那个时候需要用到这个map(反查询)
	ep.tables.Store(conn.id, conn)
	conn.BindEpoller(e) // 绑定epoller
	return nil
}
func (e *epoller) remove(c *connection) error {
	subTcpNum()
	fd := c.fd
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	ep.tables.Delete(c.id)
	e.fdToConnTable.Delete(c.fd)
	return nil
}

// EpollWait 会阻塞直到有事件发生，返回发生事件的fd
func (e *epoller) wait(msec int) ([]*connection, error) {
	events := make([]unix.EpollEvent, config.GetGatewayEpollWaitQueueSize())
	n, err := unix.EpollWait(e.fd, events, msec)
	if err != nil {
		return nil, err
	}
	var connections []*connection
	for i := 0; i < n; i++ {
		// epoll在wait时只能拿到fd，需要通过fd找到对应的connection
		if conn, ok := e.fdToConnTable.Load(int(events[i].Fd)); ok { // epoll只能知道每个触发事件的fd，需要通过fd找到对应的connection
			connections = append(connections, conn.(*connection))
		}
	}
	return connections, nil
}
func socketFD(conn *net.TCPConn) int {
	file, err := conn.File()
	if err != nil {
		panic(err)
	}
	fd := int(file.Fd())
	return fd
	// tcpConn := reflect.Indirect(reflect.ValueOf(*conn)).FieldByName("conn")
	// fdVal := tcpConn.FieldByName("fd")
	// pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	// return int(pfdVal.FieldByName("Sysfd").Int())
}

// 设置go 进程打开文件数的限制
func setLimit() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	log.Printf("set cur limit: %d", rLimit.Cur)
}

func addTcpNum() {
	atomic.AddInt32(&tcpNum, 1)
}

func getTcpNum() int32 {
	return atomic.LoadInt32(&tcpNum)
}
func subTcpNum() {
	atomic.AddInt32(&tcpNum, -1)
}

func checkTcp() bool {
	num := getTcpNum()
	maxTcpNum := config.GetGatewayMaxTcpNum() // 70000
	return num <= maxTcpNum
}

func setTcpConifg(c *net.TCPConn) {
	_ = c.SetKeepAlive(true)
}
