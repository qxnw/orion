package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
)

//Response 处理api返回值
func Response(conf *conf.MetadataConf) dispatcher.HandlerFunc {
	return func(ctx *dispatcher.Context) {
		ctx.Next()
		nctx := getCTX(ctx)
		if nctx == nil {
			return
		}
		defer nctx.Close()
		if err := nctx.Response.GetError(); err != nil {
			getLogger(ctx).Errorf("err:%v", err)
			if !servers.IsDebug {
				nctx.Response.ShouldContent(errors.New("请求发生错误"))
			}
		}
		if ctx.Writer.Written() {
			return
		}
		switch nctx.Response.GetContentType() {
		case 1:
			ctx.SecureJSON(nctx.Response.GetStatus(), nctx.Response.GetContent())
		case 2:
			ctx.XML(nctx.Response.GetStatus(), nctx.Response.GetContent())
		default:
			if content, ok := nctx.Response.GetContent().(string); ok {
				if (strings.HasPrefix(content, "[") || strings.HasPrefix(content, "{")) &&
					(strings.HasSuffix(content, "}") || strings.HasSuffix(content, "]")) {
					ctx.SecureJSON(nctx.Response.GetStatus(), nctx.Response.GetContent())
				} else {
					ctx.Data(nctx.Response.GetStatus(), "text/plain", []byte(nctx.Response.GetContent().(string)))
				}
				return
			}
			ctx.Data(nctx.Response.GetStatus(), "text/plain", []byte(fmt.Sprint(nctx.Response.GetContent())))
		}
	}
}
