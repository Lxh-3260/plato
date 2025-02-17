package source

import (
	"context"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/lxh-3260/plato/common/config"
	"github.com/lxh-3260/plato/common/discovery"
)

func Init() {
	eventChan = make(chan *Event)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 优雅地关闭子协程

	go DataHandler(ctx)
	if config.IsDebug() {
		ctx := context.Background()
		testServiceRegister(&ctx, "7896", "node1")
		testServiceRegister(&ctx, "7897", "node2")
		testServiceRegister(&ctx, "7898", "node3")
	}
}

// 服务发现处理
func DataHandler(ctx context.Context) {
	dis := discovery.NewServiceDiscovery(ctx)
	defer dis.Close()
	setFunc := func(key, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); ed != nil {
				event.Type = AddNodeEvent
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(ctx, "DataHandler.setFunc.err :%s", err.Error())
		}
	}
	delFunc := func(key, value string) {
		if ed, err := discovery.UnMarshal([]byte(value)); err == nil {
			if event := NewEvent(ed); ed != nil {
				event.Type = DelNodeEvent // 没有直接在这里删除，而是向channel中发一个delete时间，由dispatcher.go中的goroutine处理
				eventChan <- event
			}
		} else {
			logger.CtxErrorf(ctx, "DataHandler.delFunc.err :%s", err.Error())
		}
	}
	err := dis.WatchService(config.GetServicePathForIPConf(), setFunc, delFunc) // 在mock中key为"/plato/ip_dispatcher/node1 2 3"，后续通过etcd的watcher监控服务变化（通过前缀找到对应的集合）
	if err != nil {
		panic(err)
	}
}
