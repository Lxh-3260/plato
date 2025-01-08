package domain

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

type IpConfConext struct {
	Ctx       *context.Context    // 协程上下文控制
	AppCtx    *app.RequestContext // HTTP请求上下文控制
	ClinetCtx *ClientConext
}

type ClientConext struct {
	IP string `json:"ip"`
}

func BuildIpConfContext(c *context.Context, ctx *app.RequestContext) *IpConfConext {
	ipConfConext := &IpConfConext{
		Ctx:       c,
		AppCtx:    ctx,
		ClinetCtx: &ClientConext{},
	}
	return ipConfConext
}
