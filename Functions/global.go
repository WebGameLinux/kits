package Functions

import (
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Supports"
		"reflect"
)

// 获取容器中的服务
func App(key string) interface{} {
		var obj = Supports.App().Get(key)
		if obj == nil {
				return nil
		}
		return obj
}

// 配置服务
func Config() Components.GetterInterface {
		var config = App("config")
		if config == nil {
				return nil
		}
		if conf, ok := config.(Components.GetterInterface); ok {
				return conf
		}
		return nil
}

// 获取配置
func CnfKv(key string, defaults ...interface{}) interface{} {
		return Config().Any(key, defaults...)
}

// 容器服务
func AppContainer() Contracts.ApplicationContainer {
		return App("app").(Contracts.ApplicationContainer)
}

// 是否实现某个接口
// obj any
// face new(Interface)
func Implements(obj interface{}, face interface{}) bool {
		t := reflect.TypeOf(obj)
		if t.Kind() != reflect.Interface && t.Kind() != reflect.Ptr {
				return false
		}
		return t.Implements(reflect.TypeOf(face).Elem())
}
