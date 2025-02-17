package tcp

import (
	"bytes"
	"encoding/binary"
)

type DataPgk struct {
	Len  uint32
	Data []byte
}

func (d *DataPgk) Marshal() []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, d.Len) // 先写入 4 字节长度
	return append(bytesBuffer.Bytes(), d.Data...)      // 再追加数据
}
