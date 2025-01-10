package perf

import (
	"net"

	"github.com/lxh-3260/plato/common/sdk"
)

var (
	TcpConnNum int32
)

// 压测入口
func RunMain() {
	for i := 0; i < int(TcpConnNum); i++ {
		sdk.NewChat(net.ParseIP("127.0.0.1"), 8900, "xinghao", "1111", "11111")
	}
}
