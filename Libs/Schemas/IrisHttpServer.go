package Schemas

import (
		"fmt"
		"github.com/kataras/iris"
		"github.com/kataras/iris/core/host"
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"strings"
		"sync"
)

type irisHttpServer struct {
		Name       string
		irisServer *iris.Application
		bean       Contracts.SupportInterface
		app        Contracts.ApplicationContainer
		clazz      Contracts.ClazzInterface
		runner     iris.Runner
}

type PreparesFunc func(app *iris.Application)
type RegisterFunc func(app Contracts.ApplicationContainer)

type IrisConfigureGetter func() iris.Configuration
type IrisConfigureLoader func(app Contracts.ApplicationContainer) iris.Configuration

// 配置获取器
type IrisConfigureProviderInterface interface {
		GetConfig(Contracts.ApplicationContainer) iris.Configuration
}

type IrisHttpServerProvider interface {
		Server() *iris.Application
}

var (
		irisInstanceLock  sync.Once
		irisApp           *irisHttpServer
		defaultInjectTags = []string{"json", "toml", "yml"}
)

const (
		IrisHttpServerClass                   = "IrisHttpServer"
		IrisApplication                       = "IrisApplication"
		IrisAppState                          = "IrisInstanceState"
		IrisConfigurationProvider             = "IrisConfigurationProvider"
		IrisRegisterAfters                    = "IrisRegisterAfters"
		IrisRunnerHostConfigurators           = "IrisRunnerHostConfigurators"
		IrisConfigPrefixKey                   = "http.iris"
		IrisRunner                            = "IrisRunner"
		IrisConfigurationProviderBootPrepares = "IrisConfigurationProviderBootPrepares"
)

func irisHttpServerNew() {
		irisApp = new(irisHttpServer)
		irisApp.irisServer = iris.New()
		irisApp.Name = IrisHttpServerClass
}

func IrisHttpServerOf() *irisHttpServer {
		if irisApp == nil {
				irisInstanceLock.Do(irisHttpServerNew)
		}
		return irisApp
}

func (this *irisHttpServer) Init(app Contracts.ApplicationContainer) {
		if this.app == nil {
				this.app = app
		}
}

func (this *irisHttpServer) Register() {
		if !this.app.Exists(this.String()) {
				this.app.Bind(this.String(), this)
		}
		if !this.app.Exists(IrisApplication) {
				this.app.Singleton(IrisApplication, this.getIrisAppInstance)
		}
		// iris 服务注册之后
		this.registerAfter()
}

// 注册之后
func (this *irisHttpServer) registerAfter() {
		registerAfters := this.app.Get(IrisRegisterAfters)
		if registerAfters == nil {
				return
		}
		if arr, ok := registerAfters.([]func(Contracts.ApplicationContainer)); ok {
				for _, fn := range arr {
						fn(this.app)
				}
				return
		}
		if fn, ok := registerAfters.(func(Contracts.ApplicationContainer)); ok {
				fn(this.app)
				return
		}
		if fn, ok := registerAfters.(func(provider IrisHttpServerProvider)); ok {
				fn(this)
				return
		}
}

func (this *irisHttpServer) Server() *iris.Application {
		return this.GetServer()
}

// 配置注入
func (this *irisHttpServer) getIrisAppInstance(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		this.Server().Configure(this.getIrisConfigure())
		return this.Server
}

func (this *irisHttpServer) getIrisConfigure() iris.Configurator {
		return iris.WithConfiguration(this.configure())
}

// 获取iris配置
func (this *irisHttpServer) configure() iris.Configuration {
		provider := this.app.Get(IrisConfigurationProvider)
		if provider != nil {
				if cfn, ok := provider.(IrisConfigureGetter); ok {
						return cfn()
				}
				if cfn, ok := provider.(func() iris.Configuration); ok {
						return cfn()
				}
				if loader, ok := provider.(IrisConfigureLoader); ok {
						return loader(this.app)
				}
				if loader, ok := provider.(func(app Contracts.ApplicationContainer) iris.Configuration); ok {
						return loader(this.app)
				}
				if pro, ok := provider.(IrisConfigureProviderInterface); ok {
						return pro.GetConfig(this.app)
				}
				if config, ok := provider.(iris.Configuration); ok {
						return config
				}
				if config, ok := provider.(*iris.Configuration); ok {
						return *config
				}
		}
		return this.config()
}

// 由服务提供注入
func (this *irisHttpServer) config() iris.Configuration {
		var cnf = iris.DefaultConfiguration()
		// 注入失败返回默认配置
		if !this.getConfigureProvider().Inject(IrisConfigPrefixKey, &cnf, defaultInjectTags...) {
				return iris.DefaultConfiguration()
		}
		return cnf
}

// 获取配置服务
func (this *irisHttpServer) getConfigureProvider() Components.ConfigureProvider {
		var config = this.app.Get(Components.ConfigureProviderClass)
		if provider, ok := config.(Components.ConfigureProvider); ok {
				return provider
		}
		return Components.ConfigureProviderOf()
}

// boot 引导启动
func (this *irisHttpServer) Boot() {
		this.prepare()
		this.StartUp()
}

// boot 启动前置
func (this *irisHttpServer) prepare() {
		// 注入配置
		this.Server().Configure(this.getIrisConfigure())
		// 获取前置 逻辑
		bootPrepares := this.app.Get(IrisConfigurationProviderBootPrepares)
		if bootPrepares == nil {
				return
		}
		// 服务启动前置组
		if prepares, ok := bootPrepares.([]func(*iris.Application)); ok {
				for _, fn := range prepares {
						fn(this.GetServer())
				}
		}
		// 单个
		fn, ok := bootPrepares.(func(*iris.Application))
		if ok {
				fn(this.GetServer())
		}
}

func (this *irisHttpServer) StartUp() {
		if this.started() {
				return
		}
		go this.run()
		this.start()
}

// 启动服务
func (this *irisHttpServer) run() {
		err := this.Server().Run(this.getServerRunner())
		this.logger(err)
		if err != nil {
				this.app.Stop()
		}
}

// 服务停止日志记录
func (this *irisHttpServer) logger(stringer interface{}) {
		if stringer == nil {
				return
		}
		if str, ok := stringer.(string); ok {
				this.Server().Logger().Println(str)
		}
		if str, ok := stringer.([]interface{}); ok {
				this.Server().Logger().Println(str...)
		}
		if str, ok := stringer.([]string); ok {
				this.Server().Logger().Println(strings.Join(str, " "))
		}
		if str, ok := stringer.(fmt.Stringer); ok {
				this.Server().Logger().Println(str.String())
		}
		if err, ok := stringer.(error); ok {
				this.Server().Logger().Error(err.Error())
		}
}

func (this *irisHttpServer) getServerRunner() iris.Runner {
		if this.runner == nil {
				runner := this.app.Get(IrisRunner)
				if fn, ok := runner.(iris.Runner); ok {
						this.runner = fn
				} else {
						this.runner = iris.Addr(this.GetHttpAddr(), this.getConfigurator()...)
				}
		}
		return this.runner
}

// 获取runner host configurator
func (this *irisHttpServer) getConfigurator() []host.Configurator {
		items := this.app.Get(IrisRunnerHostConfigurators)
		if fn, ok := items.(host.Configurator); ok {
				return []host.Configurator{fn}
		}
		if fn, ok := items.(func(*iris.Supervisor)); ok {
				return []host.Configurator{fn}
		}
		if arr, ok := items.([]func(*iris.Supervisor)); ok {
				var all []host.Configurator
				for _, it := range arr {
						all = append(all, it)
				}
				return all
		}
		if arr, ok := items.([]host.Configurator); ok {
				return arr
		}
		return []host.Configurator{}
}

func (this *irisHttpServer) GetHttpAddr() string {
		addr := this.getConfigureProvider().Get(Contracts.HttpAddrConfig)
		if addr != "" {
				return addr
		}
		_host := this.getConfigureProvider().Get(Contracts.HttpHostConfig)
		_port := this.getConfigureProvider().Get(Contracts.HttpPortConfig, Contracts.HttpPortDefault)
		return fmt.Sprintf("%s:%s", _host, _port)
}

func (this *irisHttpServer) started() bool {
		state := this.app.Get(IrisAppState)
		if state == nil {
				return false
		}
		if n, ok := state.(int); ok {
				return n > 0
		}
		if n, ok := state.(bool); ok {
				return n
		}
		return false
}

func (this *irisHttpServer) start() {
		this.app.Bind(IrisAppState, 1)
}

func (this *irisHttpServer) stop() {
		this.app.Bind(IrisAppState, -1)
}

func (this *irisHttpServer) GetSupportBean() Contracts.SupportInterface {
		if this.bean == nil {
				this.bean = Components.BeanOf()
		}
		return this.bean
}

func (this *irisHttpServer) Factory(container Contracts.ApplicationContainer) interface{} {
		this.Init(container)
		return this
}

func (this *irisHttpServer) Constructor() interface{} {
		return IrisHttpServerOf()
}

func (this *irisHttpServer) GetServer() *iris.Application {
		return this.irisServer
}

func (this *irisHttpServer) Static(requestPath string, systemPath string) interface{} {
		return this.Server().StaticWeb(requestPath, systemPath)
}

func (this *irisHttpServer) String() string {
		return this.Name
}

func (this *irisHttpServer) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = Components.ClazzOf(this)
		}
		return this.clazz
}
