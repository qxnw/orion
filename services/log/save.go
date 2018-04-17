package log

import (
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/orion/modules/elastic"
	"github.com/qxnw/orion/modules/logging"
)

type SaveHandler struct {
	container component.IContainer
	name      string
}

func NewSaveHandlerByName(n string) func(container component.IContainer) (u *SaveHandler) {
	return func(container component.IContainer) (u *SaveHandler) {
		return &SaveHandler{
			container: container,
			name:      n,
		}
	}
}

//NewSaveHandler 创建服务
func NewSaveHandler(container component.IContainer) (u *SaveHandler) {
	return &SaveHandler{
		container: container,
	}
}

//Handle 保存日志记录
func (u *SaveHandler) Handle(name string, engine string, service string, ctx *context.Context) (r interface{}) {
	ctx.Log.Info("--------保存日志----------")
	body, err := ctx.Request.Ext.GetBody()
	if err != nil {
		return err
	}
	if len(body) <= 2 {
		ctx.Response.SetStatus(204)
		return nil
	}
	logger, err := u.container.GetGlobalObject(elastic.ConfNode, u.name)
	if err != nil {
		return err
	}
	logging := logger.(*logging.LoggingService)
	if err = logging.Save(body); err != nil {
		return err
	}
	return "success"
}
