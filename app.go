package main

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/hydra/hydra"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/orion/modules/elastic"
	"github.com/qxnw/orion/modules/logging"
	"github.com/qxnw/orion/services/log"
)

//AppConf 应用程序全局配置
type AppConf struct {
	Names []string `json:"dbs" valid:"ascii,required"`
}

//bindConf 绑定启动配置， 启动时检查注册中心配置是否存在，不存在则引导用户输入配置参数并自动创建到注册中心
func bindConf(app *hydra.MicroApp) {
	app.Binder.API.SetMainConf(`{"address":"#address"}`)
	app.Binder.API.SetSubConf("app", `{"dbs":["hydra"]}`)
	app.Binder.RPC.SetMainConf(`{"address":"#address"}`)
	app.Binder.RPC.SetSubConf("app", `{"dbs":["hydra"]}`)
	app.Binder.Plat.SetVarConf("elastic", "hydra", `{
		"address": "#address",
		"index": "logging",		
		"write-timeout": 50,
		"cron": "@every 10s"
	}`)
}

//bind 绑定应用程序的全局变量, 根据配置的日志数据库,创建日志保存对象,注册日志服务
func bind(r *hydra.MicroApp) {
	bindConf(r)
	r.Initializing(func(c component.IContainer) error {
		var config AppConf
		if err := c.GetAppConf(&config); err != nil {
			return err
		}
		if b, err := govalidator.ValidateStruct(&config); !b {
			err = fmt.Errorf("app 配置文件有误:%v", err)
			return err
		}
		if len(config.Names) == 0 {
			err := fmt.Errorf("未配置日志名称")
			return err
		}
		for _, name := range config.Names {
			_, _, err := c.SaveGlobalObject(elastic.ConfNode, name, func(cn conf.IConf) (interface{}, error) {
				config, err := elastic.GetConf(cn)
				if err != nil {
					return nil, err
				}
				config.TypeName = name
				client, err := elastic.GetClient(c, cn)
				if err != nil {
					return nil, err
				}
				return logging.NewLoggingService(client, config, logger.GetSession(c.GetServerName(), logger.CreateSession()))
			})
			if err != nil {
				return err
			}
			r.Micro(fmt.Sprintf("/%s/log/save", name), log.NewSaveHandlerByName(name)) //根据配置的日志名称，初始化服务
		}
		return nil
	})
}
