package middleware

import (
	"fmt"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//NoResponse 处理无响应的返回结果
func NoResponse(conf *conf.MetadataConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		context := getCTX(ctx)
		if context == nil {
			return
		}
		defer context.Close()
		if err := context.Response.GetError(); err != nil {
			getLogger(ctx).Error(err)
		}
		if ctx.Writer.Written() {
			return
		}
		ctx.Writer.WriteHeader(context.Response.GetStatus())
		ctx.Writer.WriteString(fmt.Sprint(context.Response.GetContent()))
	}
}
