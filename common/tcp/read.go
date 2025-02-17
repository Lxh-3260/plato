package tcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

// 封装tcp读写和编码，gateway层不需要感知消息的内容，直接读取消息的内容
func ReadData(conn *net.TCPConn) ([]byte, error) {
	var dataLen uint32
	dataLenBuf := make([]byte, 4)
	if err := readFixedData(conn, dataLenBuf); err != nil {
		return nil, err
	}
	// fmt.Printf("readFixedData:%+v\n", dataLenBuf)
	buffer := bytes.NewBuffer(dataLenBuf)
	if err := binary.Read(buffer, binary.BigEndian, &dataLen); err != nil { // 大端读取
		return nil, fmt.Errorf("read headlen error:%s", err.Error())
	}
	if dataLen <= 0 {
		return nil, fmt.Errorf("wrong headlen :%d", dataLen)
	}
	dataBuf := make([]byte, dataLen)
	// fmt.Printf("readFixedData.dataLen:%+v\n", dataLen)
	if err := readFixedData(conn, dataBuf); err != nil {
		return nil, fmt.Errorf("read headlen error:%s", err.Error())
	}
	return dataBuf, nil
}

// 读取固定buf长度的数据
func readFixedData(conn *net.TCPConn, buf []byte) error {
	_ = (*conn).SetReadDeadline(time.Now().Add(time.Duration(120) * time.Second)) // 120s还没读取到就超时，解决tcp粘包问题（此处是假设tcp包较小，粘包较小）
	// TODO: 后续可以改进为非阻塞的方式，如果出现丢包粘包，向epoll注册一个读事件，等待下次读取
	var pos int = 0
	var totalSize int = len(buf)
	for {
		c, err := (*conn).Read(buf[pos:])
		if err != nil {
			return err
		}
		pos = pos + c
		if pos == totalSize { // 直到读够 totalSize，这样即使 TCP 拆包，readFixedData() 也能拼接完整数据包。
			break
		}
	}
	return nil
}
