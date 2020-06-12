package Supports

import (
		"fmt"
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"reflect"
		"sync"
		"time"
)

const (
		iocLock                = "ioc"
		bootName               = "boot"
		AppContainer           = "app"
		appSingletonLock       = "appSingleton"
		providerName           = "provider"
		registerName           = "register"
		propertiesLock         = "properties"
		providerLock           = "provider"
		registersPropKey       = "registers"
		bootsPropKey           = "boots"
		defaultPropsLock       = "defaultProps"
		coreRegistersInitCount = "core_registers_count"
		coreBootInitCount      = "core_boots_count"
		stateKeyTpl            = "init_%s_state"
		defaultPropsKey        = "defaultProperties"
		// userPropsKey           = "userProperties"
		coreProviderNum = "coreProviderNum"
		ctrlChan        = "appCtrlChan"
		BasePath        = "BasePath"
		StartEv         = "started"
		StopEv          = "stoped"
)

var (
		instanceMutex    = sync.Mutex{}
		appInstanceLocks = map[string]*sync.Once{
				iocLock:          getInstanceLock(),
				propertiesLock:   getInstanceLock(),
				providerLock:     getInstanceLock(),
				defaultPropsLock: getInstanceLock(),
				appSingletonLock: getInstanceLock(),
		}
		appSingleton *ApplicationImpl
)

type ApplicationImpl struct {
		properties *sync.Map
		container  ContainerApp
		registers  RegisterUniqueArray
		boots      BooterUniqueArray
}

// 获取并发单例锁
func getInstanceLock() *sync.Once {
		return &sync.Once{}
}

// 获取安全锁
func getSafeLock(key string) *sync.Once {
		var (
				ok   bool
				lock *sync.Once
		)
		instanceMutex.Lock()
		defer instanceMutex.Unlock()
		if appInstanceLocks == nil {
				appInstanceLocks = make(map[string]*sync.Once)
		}
		if lock, ok = appInstanceLocks[key]; ok {
				return lock
		}
		appInstanceLocks[key] = getInstanceLock()
		return lock
}

// 获取单例 AppContainer
func App() Contracts.ApplicationContainer {
		if appSingleton == nil {
				getSafeLock(appSingletonLock).Do(appFactory)
		}
		return appSingleton
}

// app 工厂
func appFactory() {
		appSingleton = NewApp()
		appSingleton.InitFn()
}

// 空数据构造
func InstanceOfConstructor(v interface{}) func() interface{} {
		if v == nil {
				return nil
		}
		if en, ok := v.(*entry); ok {
				clazz, ok := en.extras.Load(REAL_CLAZZ)
				if ok && clazz != nil {
						v = clazz
				} else {
						v = en.value
				}
		}
		if fn, ok := v.(func() interface{}); ok {
				return fn
		}
		if obj, ok := v.(Contracts.ClazzInterface); ok {
				fn := obj.Constructor()
				if fn != nil {
						return fn
				}
		}
		return nil
}

// 注入构造
func InstanceOfFactory(v interface{}) func(app Contracts.ApplicationContainer) interface{} {
		if v == nil {
				return nil
		}
		if en, ok := v.(*entry); ok {
				clazz, ok := en.extras.Load(REAL_CLAZZ)
				if ok && clazz != nil {
						v = clazz
				} else {
						v = en.value
				}
		}
		if fn, ok := v.(func(Contracts.ApplicationContainer) interface{}); ok {
				return fn
		}
		if obj, ok := v.(Contracts.ClazzInterface); ok {
				fn := obj.Factory()
				if fn != nil {
						return fn
				}
		}
		return nil
}

// 创建App
func NewApp() *ApplicationImpl {
		var app = new(ApplicationImpl)
		app.boots = BooterUniqueArrayOf()
		app.registers = RegisterUniqueArrayOf()
		return app
}

// 初始化函数
func (this *ApplicationImpl) InitFn() {
		if this.container == nil {
				this.IocInit()
		}
		if this.properties == nil {
				this.PropsInit()
		}
		if !this.isInit(providerName) {
				this.InitCoreProviders()
		}
}

func (this *ApplicationImpl) stateKey(key string) string {
		return fmt.Sprintf(stateKeyTpl, key)
}

// 检查对应初始化状态
func (this *ApplicationImpl) isInit(key string) bool {
		if state, ok := this.properties.Load(this.stateKey(key)); ok {
				if b, ok := state.(bool); ok {
						return b
				}
		}
		return false
}

// 设置初始化状态
func (this *ApplicationImpl) setInit(key string, state ...bool) *ApplicationImpl {
		if len(state) == 0 {
				state = append(state, true)
		}
		this.properties.Store(this.stateKey(key), state[0])
		return this
}

// ioc 容器初始化
func (this *ApplicationImpl) IocInit() {
		getSafeLock(iocLock).Do(this.iocInitFactory)
}

// app 相关属性初始化
func (this *ApplicationImpl) PropsInit() {
		getSafeLock(propertiesLock).Do(this.propertiesInitFactory)
}

// 初始化核心 服务器提供器
func (this *ApplicationImpl) InitCoreProviders() {
		getSafeLock(providerLock).Do(this.initDefaultProviders)
}

// 注册默认服务
func (this *ApplicationImpl) initDefaultProviders() {
		v, ok := this.properties.Load(defaultPropsKey)
		if !ok {
				return
		}
		if props, ok := v.(*ApplicationProps); ok {
				for _, provider := range props.Providers {
						this.Register(provider)
				}
		}
}

// 获取默认属性
func (this *ApplicationImpl) getDefaultProps() *ApplicationProps {
		props, ok := this.properties.Load(defaultPropsKey)
		if ok {
				return props.(*ApplicationProps)
		}
		return nil
}

// 获取相关初始统计数值
func (this *ApplicationImpl) getInitCount(key string) int {
		v := this.property(key, 0)
		return v.(int)
}

// 初始化相关注册器
func (this *ApplicationImpl) InitRegisters() {
		if this.getInitCount(coreRegistersInitCount) > 0 {
				return
		}
		this.registers.Foreach(this.foreachRegister())
}

// 别名
func (this *ApplicationImpl) Alias(clazz string, alias string) {
		this.container.Alias(clazz, alias)
}

// 初始化相关引导器
func (this *ApplicationImpl) InitBoots() {
		if this.getInitCount(coreBootInitCount) > 0 {
				return
		}
		this.boots.Foreach(this.foreachBoot())
}

// 载入引导逻辑
func (this *ApplicationImpl) boot(impl Contracts.BootInterface) {
		impl.Boot()
		this.properties.Store(coreBootInitCount, this.getInitCount(coreBootInitCount)+1)
}

// 载入注册逻辑
func (this *ApplicationImpl) reg(impl Contracts.RegisterInterface) {
		impl.Register()
		this.properties.Store(coreRegistersInitCount, this.getInitCount(coreRegistersInitCount)+1)
}

// 属性加载器
func (this *ApplicationImpl) PropertyLoader(loader func(p *sync.Map)) {
		if loader != nil {
				loader(this.properties)
		}
}

// property
func (this *ApplicationImpl) property(key string, defaultValue ...interface{}) interface{} {
		var (
				ok    bool
				value interface{}
		)
		if len(defaultValue) == 0 {
				defaultValue = append(defaultValue, nil)
		}
		if value, ok = this.properties.Load(key); ok {
				if defaultValue[0] == nil {
						return value
				}
				if reflect.TypeOf(defaultValue[0]) == reflect.TypeOf(value) {
						return value
				}
				return defaultValue[0]
		}
		return defaultValue[0]
}

// 获取相关服务或者状态
func (this *ApplicationImpl) Get(faced string) interface{} {
		entry := this.container.Resolver(faced)
		if entry == nil {
				return nil
		}
		if !entry.Extras().Bool(SINGLETON) {
				class, ok := entry.Extras().Load(REAL_CLAZZ)
				if ok && class != nil {
						return class
				}
				return entry.value
		}
		if obj, ok := entry.Extras().Load(SINGLETON_OBJECT); ok {
				return obj
		}
		constructor := InstanceOfConstructor(entry.value)
		if constructor != nil {
				instance := constructor()
				if instance != nil {
						entry.Extras().Store(SINGLETON_OBJECT, instance)
				}
				return instance
		}
		factory := InstanceOfFactory(entry.value)
		if factory != nil {
				instance := factory(this)
				if instance != nil {
						entry.Extras().Store(SINGLETON_OBJECT, instance)
				}
				return instance
		}
		return nil
}

// 获取相关服务或者状态
func (this *ApplicationImpl) Exists(faced string) bool {
		return this.container.Exists(faced)
}

// 注册服务提供器
func (this *ApplicationImpl) Register(provider Contracts.Provider) {
		if provider == nil {
				return
		}
		this.register(provider)
}

// 载入核心服务提供集合
func (this *ApplicationImpl) loadCoreProviders() {
		var (
				i         int
				provider  Contracts.Provider
				providers = this.getCoreProviders()
		)
		if providers == nil || len(providers) == 0 {
				return
		}
		for i, provider = range providers {
				this.register(provider)
		}
		this.properties.Store(coreProviderNum, i+1)
}

// 获取核心服务提供
func (this *ApplicationImpl) getCoreProviders() []Contracts.Provider {
		var providers []Contracts.Provider
		v := this.getDefaultProps()
		if v == nil {
				return providers
		}
		return v.GetProviders()
}

// 注册服务提供
func (this *ApplicationImpl) register(provider Contracts.Provider) {
		if provider == nil {
				return
		}
		// 注入 app
		provider.Init(this)
		// 提供相关注册选择
		var bean = provider.GetSupportBean()
		if bean.HasBoot() {
				this.boots.Add(provider)
		}
		if bean.HasRegister() {
				this.registers.Add(provider)
		}
}

// 对象绑定
func (this *ApplicationImpl) Bind(id string, object interface{}) {
		if object == nil || id == "" {
				return
		}
		this.container.Bind(id, object)
}

// 单例注册
func (this *ApplicationImpl) Singleton(id string, factory func(Contracts.ApplicationContainer) interface{}) {
		if id == "" {
				return
		}
		if this.container.Exists(id) {
				return
		}
		this.container.Singleton(id, factory)
}

// ioc 容器
func (this *ApplicationImpl) iocInitFactory() {
		if this.container == nil {
				this.container = ContainerOf()
		}
		// 注入app自身
		this.container.Bind(AppContainer, this)
}

// 属性工厂
func (this *ApplicationImpl) propertiesInitFactory() {
		var mapper = sync.Map{}
		this.properties = &mapper
		props := getApplicationDefaultProps()
		this.properties.Store(defaultPropsKey, props)
}

// 发送事件
func (this *ApplicationImpl) Emit(event string, target interface{}) {
		// @todo
		if Components.Debug() {
				fmt.Println(event, target)
		}
}

// 获取当前register加载的位置
func (this *ApplicationImpl) getRegisterPointer() int {
		return this.getInitCount(coreRegistersInitCount) - 1
}

// 获取当前boot加载的位置
func (this *ApplicationImpl) getBootPointer() int {
		return this.getInitCount(coreBootInitCount) - 1
}

// 初始服务提供
func (this *ApplicationImpl) providersInit() {
		this.LoadCoreProviders()
		this.LoadCustomProviders()
}

// 加载核心服务
func (this *ApplicationImpl) LoadCoreProviders() {
		if !this.isInit(registerName) {
				this.InitRegisters()
		}
		if !this.isInit(bootName) {
				this.InitBoots()
		}
}

// 加载自定义 服务器提供器
func (this *ApplicationImpl) LoadCustomProviders() {
		this.registers.Foreach(this.foreachRegister())
		this.boots.Foreach(this.foreachBoot())
}

// each register
func (this *ApplicationImpl) foreachRegister() func(key, value interface{}) bool {
		return func(key, value interface{}) bool {
				if reg, ok := value.(Contracts.RegisterInterface); ok {
						this.reg(reg)
				}
				this.registers.OffsetUnset(key)
				return true
		}
}

// each boot
func (this *ApplicationImpl) foreachBoot() func(key, value interface{}) bool {
		return func(key, value interface{}) bool {
				if boot, ok := value.(Contracts.BootInterface); ok {
						this.boot(boot)
				}
				this.boots.OffsetUnset(key)
				return true
		}
}

// 获取所有属性
func (this *ApplicationImpl) Profiles() map[string]interface{} {
		var profiles = make(map[string]interface{})
		if props, ok := this.properties.Load(defaultPropsKey); ok {
				if v, ok := props.(ApplicationProps); ok {
						v.Foreach(func(key string, value interface{}) bool {
								profiles[key] = value
								return true
						})
				}
		}
		this.properties.Range(func(key, value interface{}) bool {
				if k, ok := key.(string); ok {
						if k != defaultPropsKey {
								profiles[k] = value
						}
				}
				return true
		})
		return profiles
}

// 获取属性
func (this *ApplicationImpl) GetProfile(key string) interface{} {
		if v, ok := this.properties.Load(key); ok {
				return v
		}
		if v := this.getDefaultProps(); v != nil {
				return v.Get(key)
		}
		return nil
}

// 启动服务器监听逻辑
func (this *ApplicationImpl) StarUp() {
		var (
				ok bool
				ch chan int
		)
		v := this.GetProfile(ctrlChan)
		// 是否运行
		stop := this.GetProfile(HelpStop)
		if s, ok := stop.(bool); ok && s {
				this.Stop()
				return
		}

		if v == nil {
				v = make(chan int, 2)
		}
		if ch, ok = v.(chan int); ok {
				this.properties.Store(ctrlChan, ch)
		}
		//  providers
		this.providersInit()
		this.Emit(StartEv, ch)
		ticker := time.NewTicker(3 * time.Second)
		// 等待结束
		for {
				select {
				case signal := <-ch:
						if signal == -1 {
								close(ch)
								ticker.Stop()
								return
						}
				case <-ticker.C:
						health := this.Get(Contracts.AppHealth)
						if fn, ok := health.(func()); ok {
								fn()
						}
				}
		}
}

// 停止服务
func (this *ApplicationImpl) Stop() {
		ch, ok := this.properties.Load(ctrlChan)
		if !ok {
				return
		}
		if ch1, ok := ch.(chan int); ok {
				ch1 <- -1
				this.Emit(StopEv, ch)
		}
}
