package ipconf

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/lxh-3260/plato/common/config"
	"github.com/lxh-3260/plato/ipconf/domain"
	"github.com/lxh-3260/plato/ipconf/source"
)

// RunMain 启动web容器
func RunMain(path string) { // path = "./plato.yaml"
	config.Init(path)
	source.Init() //数据源要优先启动
	domain.Init() // 初始化调度层
	s := server.Default(server.WithHostPorts(":6789"))
	s.GET("/ip/list", GetIpInfoList)
	s.Spin()
}
