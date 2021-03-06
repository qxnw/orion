package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/snappy"
	"github.com/qxnw/hydra/context"
	"github.com/qxnw/hydra/servers"
	"github.com/qxnw/hydra/servers/pkg/dispatcher"
	"github.com/qxnw/lib4go/encoding"
	"github.com/qxnw/lib4go/logger"
)

func getUUID(c *dispatcher.Context) string {
	sid, ok := c.Request.GetHeader()["__hydra_sid_"]
	if !ok || sid == "" {
		return logger.CreateSession()
	}
	return sid
}
func setUUID(c *dispatcher.Context, id string) {
	c.Request.GetHeader()["__hydra_sid_"] = id
}

func setStartTime(c *dispatcher.Context) {
	c.Set("__start_time_", time.Now())
}
func setLogger(c *dispatcher.Context, l *logger.Logger) {
	c.Set("__logger_", l)
}
func getLogger(c *dispatcher.Context) *logger.Logger {
	l, _ := c.Get("__logger_")
	return l.(*logger.Logger)
}
func getExpendTime(c *dispatcher.Context) time.Duration {
	start, _ := c.Get("__start_time_")
	return time.Since(start.(time.Time))
}
func getJWTRaw(c *dispatcher.Context) interface{} {
	jwt, _ := c.Get("__jwt_")
	return jwt
}
func setJWTRaw(c *dispatcher.Context, v interface{}) {
	c.Set("__jwt_", v)
}
func getServiceName(c *dispatcher.Context) string {
	if service, ok := c.Get("__service_"); ok {
		return service.(string)
	}
	return ""
}
func setServiceName(c *dispatcher.Context, v string) {
	c.Set("__service_", v)
}
func setCTX(c *dispatcher.Context, r *context.Context) {
	c.Set("__context_", r)
}
func getCTX(c *dispatcher.Context) *context.Context {
	result, _ := c.Get("__context_")
	if result == nil {
		return nil
	}
	return result.(*context.Context)
}

//ContextHandler api请求处理程序
func ContextHandler(exhandler interface{}, name string, engine string, service string, mSetting map[string]string, ext map[string]interface{}) dispatcher.HandlerFunc {

	handler, ok := exhandler.(servers.IExecuter)
	if !ok {
		panic("不是有效的servers.IExecuter接口")
	}

	return func(c *dispatcher.Context) {
		//处理输入参数
		ctx := context.GetContext(exhandler.(context.IContainer), makeQueyStringData(c), makeFormData(c), makeParamsData(c), makeSettingData(c, mSetting), makeExtData(c, ext), getLogger(c))

		defer setServiceName(c, ctx.Request.Translate(service, false))
		defer setCTX(c, ctx)
		//调用执行引擎进行逻辑处理
		result := handler.Execute(name, engine, ctx.Request.Translate(service, false), ctx)
		if result != nil {
			ctx.Response.ShouldContent(result)
		}
		//处理错误err,5xx
		if err := ctx.Response.GetError(); err != nil {
			err = fmt.Errorf("error:%v", err)
			if !servers.IsDebug {
				err = errors.New("error:Internal Server Error")
			}
			return
		}
		//处理跳转3xx
		if url, ok := ctx.Response.IsRedirect(); ok {
			c.Redirect(ctx.Response.GetStatus(), url)
			return
		}

	}
}

func makeFormData(ctx *dispatcher.Context) InputData {
	return ctx.GetPostForm
}
func makeQueyStringData(ctx *dispatcher.Context) InputData {
	return nil
}
func makeParamsData(ctx *dispatcher.Context) InputData {
	return ctx.Params.Get
}
func makeSettingData(ctx *dispatcher.Context, m map[string]string) ParamData {
	return m
}

func makeExtData(c *dispatcher.Context, ext map[string]interface{}) map[string]interface{} {
	input := make(map[string]interface{})
	for k, v := range ext {
		input[k] = v
	}
	input["__hydra_sid_"] = getUUID(c)
	input["__method_"] = strings.ToLower(c.Request.GetMethod())
	input["__header_"] = c.Request.GetHeader()
	input["__jwt_"] = getJWTRaw(c)
	input["__func_http_request_"] = c.Request
	input["__func_http_response_"] = c.Writer
	input["__binding_"] = func(obj interface{}) error {
		buffer, err := json.Marshal(c.Request.GetForm())
		if err != nil {
			return err
		}
		return json.Unmarshal(buffer, obj)
	}
	input["__binding_with_"] = func(v interface{}, ct string) error {
		buffer, err := json.Marshal(c.Request.GetForm())
		if err != nil {
			return err
		}
		return json.Unmarshal(buffer, v)
	}
	input["__func_body_get_"] = func(ch string) (string, error) {
		if buff, ok := c.Get("__body_"); ok {
			if s, ok := buff.(string); ok {
				if v, ok := c.Request.GetHeader()["__encode_snappy_"]; ok && v == "true" {
					buff := []byte(s)
					var nbuff []byte
					nbuffer, err := snappy.Decode(nbuff, buff)
					if err != nil {
						return "", err
					}
					return string(nbuffer), nil
				}
				return s, nil
			}
			return encoding.Convert(buff.([]byte), ch)
		}
		return "", errors.New("body读取错误")

	}
	input["__get_request_values_"] = func() map[string]string {
		return c.Request.GetForm()
	}
	return input
}

//InputData 输入参数
type InputData func(key string) (string, bool)

//Get 获取指定键对应的数据
func (i InputData) Get(key string) (string, error) {
	if i != nil {
		if r, ok := i(key); ok {
			return r, nil
		}
	}

	return "", fmt.Errorf("数据不存在:%s", key)
}

//ParamData map参数数据
type ParamData map[string]string

//Get 获取指定键对应的数据
func (i ParamData) Get(key string) (string, error) {
	if r, ok := i[key]; ok {
		return r, nil
	}
	return "", fmt.Errorf("数据不存在:%s", key)
}
