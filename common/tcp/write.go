package tcp

import "net"

// 将指定数据包发送到网络中
func SendData(conn *net.TCPConn, data []byte) error {
	totalLen := len(data)
	writeLen := 0
	for {
		len, err := conn.Write(data[writeLen:])
		if err != nil {
			return err
		}
		writeLen = writeLen + len
		if writeLen >= totalLen { //循环 Write() 直到全部数据发出，这样可能可以防止数据只发送了一部分，导致接收端数据不完整。
			break
		}
	}
	return nil
}
