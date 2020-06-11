package Functions

import (
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Supports"
)

// 获取容器中的服务
func App(key ...string) interface{} {
		if len(key) == 0 {
				return Supports.App()
		}
		var obj = Supports.App().Get(key[0])
		if obj == nil {
				return nil
		}
		return obj
}

// 获取配置
func CnfKv(key string, defaults ...interface{}) interface{} {
		return Config().Any(key, defaults...)
}

// 容器服务
func AppContainer() Contracts.ApplicationContainer {
		return App("app").(Contracts.ApplicationContainer)
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

// 获取环境
func Env(key string, defaults ...interface{}) interface{} {
		if len(defaults) == 0 {
				defaults = append(defaults, nil)
		}
		env := Environment()
		if env == nil {
				return nil
		}
		vars := env.Get(key)
		if vars == "" {
				return defaults[0]
		}
		return vars
}

// 环境对象
func Environment() Components.EnvironmentProvider {
		return App(Components.EnvironmentProviderClass).(Components.EnvironmentProvider)
}

