package source

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lxh-3260/plato/common/config"
	"github.com/lxh-3260/plato/common/discovery"
)

func testServiceRegister(ctx *context.Context, port, node string) {
	// 模拟服务发现
	go func() {
		//创建endpoint信息或更新endpoint负载
		createOrUpdateEndpointInfo := func() discovery.EndpointInfo {
			return discovery.EndpointInfo{
				IP:   "127.0.0.1",
				Port: port,
				MetaData: map[string]interface{}{ // nodeName to information(为了负载均衡)
					"connect_num":   float64(rand.Int63n(12312321231231131)),
					"message_bytes": float64(rand.Int63n(1231232131556)),
				},
			}
		}

		ed := createOrUpdateEndpointInfo()
		sr, err := discovery.NewServiceRegister(ctx, fmt.Sprintf("%s/%s", config.GetServicePathForIPConf(), node), &ed, time.Now().Unix())
		if err != nil {
			// panic(err) // 生产环境不要用panic，记录错误后方便排查
			log.Println("Failed to register service: ", err)
			return
		}
		go sr.ListenLeaseRespChan()
		for { // 保证服务的元数据是动态更新，用于负载均衡，无限循环来定期更新服务注册信息。
			ed = createOrUpdateEndpointInfo()
			sr.UpdateValue(&ed)         // 调用etcd.Put()更新服务注册信息，DataHandler收到通知后调用setFunc()，向eventChan发一个addNodeEvent，endpoint协程收到该通知，用atomic更新服务发现信息
			time.Sleep(1 * time.Second) // 1s更新一次每个endpoint的metadata
		}
	}()
}
