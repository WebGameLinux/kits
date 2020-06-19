package Functions

import (
		stdContext "context"
		"github.com/kataras/iris"
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Libs/Schemas"
		"github.com/webGameLinux/kits/Supports"
		"reflect"
		"time"
)

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
				app = AppContainer()
		}
		InitRegister(app)
		InitProviders(app)
		InitBootstrapper(app)
		InitAppProperties(app)
}

// 初始化 provider
func InitProviders(app Contracts.ApplicationContainer)  {
		// app.Register(Schemas.IrisHttpServerOf())
		app.Register(Components.SchemaServiceProviderOf())
		app.Register(Schemas.BeegoHttpServerOf())
}

// 初始化引导
func InitBootstrapper(app Contracts.ApplicationContainer)  {
		bootstrapper := GetBootstrap(app)
		if bootstrapper == nil {
				bootstrapper = Components.AppBootstrapperOf()
		}
		// bootstrapper.Add(Components.SchemaServiceProviderOf().(Components.Bootstrapper))
}

// 注册相关函数和对象
func InitRegister(app Contracts.ApplicationContainer)  {
		app.Bind(Components.ConfigureLoaderName, Components.ViperConfigLoader)
		app.Bind(Schemas.IrisConfigurationProviderBootPrepares, RegisterOnInterrupt)
}

// 初始化应用相关 属性配置
func InitAppProperties(app Contracts.ApplicationContainer) {
		if loader, ok := app.(Contracts.PropertyLoaderInterface); ok {
				props := Supports.AppBasePropertiesOf()
				if !props.Inited() {
						props.Init()
				}
				props.Foreach(props.Configure(loader))
		}
}

// 获取 BootstrapProvider
func GetBootstrap(container Contracts.ApplicationContainer) Components.BootstrapProvider {
		boot := container.Get(Components.AppBootstrapperClass)
		if bootstrapper, ok := boot.(Components.BootstrapProvider); ok {
				return bootstrapper
		}
		return Components.AppBootstrapperOf()
}

// 注册中断监听
func RegisterOnInterrupt(app *iris.Application) {
		iris.RegisterOnInterrupt(func() {
				timeout := 5 * time.Second
				ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
				defer cancel()
				// close all hosts
				_ = app.Shutdown(ctx)
		})
}
