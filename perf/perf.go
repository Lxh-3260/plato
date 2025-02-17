package perf

import (
	"fmt"
	"net"
	"net/http"

	"github.com/lxh-3260/plato/common/sdk"
)

var (
	TcpConnNum int32 = 10000
)

// 压测入口
/*
获取 CPU 的性能分析报告
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

获取内存的性能分析报告
go tool pprof http://localhost:6060/debug/pprof/heap

获取 goroutine 的性能分析报告
go tool pprof http://localhost:6060/debug/pprof/goroutine

查看占用的所有端口
sudo lsof -i -P -n | grep plato
ps -au | grep plato
*/
func RunMain() {
	go func() {
		fmt.Println("pprof server started at :8888")
		if err := http.ListenAndServe("0.0.0.0:8888", nil); err != nil {
			fmt.Println("Failed to start pprof server:", err)
		}
	}()
	for i := 0; i < int(TcpConnNum); i++ {
		sdk.NewChat(net.ParseIP("127.0.0.1"), 8900, "xinghao", "1111", "11111")
	}
	// 阻塞
	select {}
}
