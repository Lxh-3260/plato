package client

import (
	"context"
	"fmt"
	"time"

	"github.com/lxh-3260/plato/common/config"
	"github.com/lxh-3260/plato/common/prpc"
	"github.com/lxh-3260/plato/state/rpc/service"
)

var stateClient service.StateClient

func initStateClient() {
	pCli, err := prpc.NewPClient(config.GetStateServiceName())
	if err != nil {
		panic(err)
	}
	cli, err := pCli.DialByEndPoint(config.GetGatewayStateServerEndPoint())
	if err != nil {
		panic(err)
	}
	stateClient = service.NewStateClient(cli)
}

func CancelConn(ctx *context.Context, endpoint string, connID uint64, Payload []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 100*time.Millisecond)
	stateClient.CancelConn(rpcCtx, &service.StateRequest{ // 构造一个请求，调用state server的rpc接口
		Endpoint: endpoint,
		ConnID:   connID,
		Data:     Payload,
	})
	return nil
}

func SendMsg(ctx *context.Context, endpoint string, connID uint64, Payload []byte) error {
	rpcCtx, _ := context.WithTimeout(*ctx, 500*time.Millisecond)
	// defer cancel()
	fmt.Println("sendMsg", connID, string(Payload))
	_, err := stateClient.SendMsg(rpcCtx, &service.StateRequest{
		Endpoint: endpoint,
		ConnID:   connID,  // cmdctx中的connID是新的connID
		Data:     Payload, // payload中是protoBuf编码的消息，里面存着各种信令对应的消息体结构
		/*
			顶层cmd pb结构
			message MsgCmd{
			   CmdType Type = 1;
			   bytes Payload = 2;
			}
		*/
	})
	if err != nil {
		fmt.Printf("SendMsg failed: %v\n", err)
		return err
	}
	return nil
}
