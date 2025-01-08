package ipconf

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/lxh-3260/plato/ipconf/domain"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

// GetIpInfoList API 适配应用层
// c context.Context 和 ctx *app.RequestContext 是由 HTTP 框架在请求处理过程中自动提供的
func GetIpInfoList(c context.Context, ctx *app.RequestContext) {
	defer func() {
		if err := recover(); err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"err": err})
		}
	}()
	// Step0: 构建客户请求信息
	ipConfCtx := domain.BuildIpConfContext(&c, ctx) // ipConfCtx中包含了协程生命周期的c上下文和HTTP请求的ctx上下文
	// Step1: 进行ip调度
	eds := domain.Dispatch(ipConfCtx)
	// Step2: 根据得分取top5返回
	ipConfCtx.AppCtx.JSON(consts.StatusOK, packRes(top5Endports(eds)))
}
