package main

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/qxnw/hydra/component"
	"github.com/qxnw/hydra/conf"
	"github.com/qxnw/orion/elastic"
	"github.com/qxnw/orion/logging"
	"github.com/qxnw/orion/services/log"
)

//AppConf 应用程序配置
type AppConf struct {
	Names []string `json:"logs" valid:"ascii,required"`
}

//binding 绑定应用程序的全局对错
func binding(r component.IComponentRegistry) {
	r.Initializing(func(c component.IContainer) error {
		var config AppConf
		if err := c.GetAppConf(&config); err != nil {
			return err
		}
		if b, err := govalidator.ValidateStruct(&config); !b {
			err = fmt.Errorf("app 配置文件有误:%v", err)
			return err
		}
		for _, name := range config.Names {
			_, _, err := c.SaveGlobalObject(elastic.ConfNode, name, func(cn conf.IConf) (interface{}, error) {
				client, err := elastic.GetClient(c, cn)
				if err != nil {
					return nil, err
				}
				config, err := elastic.GetConf(cn)
				if err != nil {
					return nil, err
				}
				return logging.NewLoggingService(client, config)
			})
			if err != nil {
				return err
			}
			r.Micro(fmt.Sprintf("/%s/log/save", name), log.NewSaveHandlerByName(name)) //根据配置的日志名称，初始化服务
		}
		return nil
	})
}
