package Functions

import (
		stdContext "context"
		"fmt"
		"github.com/kataras/iris"
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Libs/Schemas"
		"github.com/webGameLinux/kits/Supports"
		"reflect"
		"time"
)

func init() {
		var props = Supports.ApplicationDefaultProps()
		fmt.Println(props.GetArgs())
}

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

// 添加项目自定义引导加载器
func Bootstrap(apps ...Contracts.ApplicationContainer) {
		if len(apps) == 0 {
				apps = append(apps, AppContainer())
		}
		app := apps[0]
		if app == nil {
				return
		}
		bootstrapper := GetBootstrap(app)
		if bootstrapper == nil {
				bootstrapper = Components.AppBootstrapperOf()
		}
		// bootstrapper.Add(Components.SchemaServiceProviderOf().(Components.Bootstrapper))
		app.Register(Components.SchemaServiceProviderOf())
		app.Register(Schemas.IrisHttpServerOf())
		app.Bind(Schemas.IrisConfigurationProviderBootPrepares,RegisterOnInterrupt)

}

// 获取 BootstrapProvider
func GetBootstrap(container Contracts.ApplicationContainer) Components.BootstrapProvider {
		boot := container.Get(Components.AppBootstrapperClass)
		if bootstrapper, ok := boot.(Components.BootstrapProvider); ok {
				return bootstrapper
		}
		return nil
}

func RegisterOnInterrupt(app *iris.Application) {
		iris.RegisterOnInterrupt(func() {
				timeout := 5 * time.Second
				ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
				defer cancel()
				// close all hosts
				_ = app.Shutdown(ctx)
		})
}
