package service

import (
	context "context"
)

const (
	DelConnCmd = 1 // DelConn
	PushCmd    = 2 // push
)

type CmdContext struct {
	Ctx     *context.Context
	Cmd     int32
	ConnID  uint64
	Payload []byte
}

type Service struct {
	CmdChannel chan *CmdContext
}

func (s *Service) DelConn(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	c := context.TODO() // 创建一个新的上下文传递进去，因为原ctx在父协程结束后会被取消，防止级联取消，这里需要异步处理cmdChannel，如果被取消有可能导致cmdChannel的关闭
	s.CmdChannel <- &CmdContext{
		Ctx:    &c,
		Cmd:    DelConnCmd,
		ConnID: gr.ConnID,
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}

func (s *Service) Push(ctx context.Context, gr *GatewayRequest) (*GatewayResponse, error) {
	c := context.TODO()
	s.CmdChannel <- &CmdContext{
		Ctx: &c,
		Cmd: PushCmd,
		// FD:      gr.FD, // 为什么传递fd不行：因为gateway和state是两个进程，fd是进程内部的，不能跨进程传递
		ConnID:  gr.ConnID,
		Payload: gr.GetData(),
	}
	return &GatewayResponse{
		Code: 0,
		Msg:  "success",
	}, nil
}
