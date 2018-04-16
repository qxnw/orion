package main

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/lib4go/logger"
	"github.com/qxnw/orion/elastic"
	"github.com/qxnw/orion/logging"
	"github.com/qxnw/orion/services/log"
)

//AppConf 应用程序全局配置
type AppConf struct {
	Names []string `json:"dbs" valid:"ascii,required"`
}

//bind 绑定应用程序的全局变量, 根据配置的日志数据库,创建日志保存对象,注册日志服务
func bind(r component.IComponentRegistry) {
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
