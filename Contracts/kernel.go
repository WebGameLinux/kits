package Contracts

import "fmt"

type Application interface {
	InitFn()
	IocInit()
	PropsInit()
	InitCoreProviders()
	InitRegisters()
	InitBoots()
	StarUp()
	Stop()
	Profiles() map[string]interface{}
	GetProfile(string) interface{}
}

type DestroyInterface interface {
	Destroy()
}

type Container interface {
	Get(string) interface{}
	Alias(string, string)
	Bind(string, interface{})
	Singleton(string, func(app ApplicationContainer) interface{})
	Destroy(...string)
	Keys() []string
	Exists(string) bool
}

type ApplicationContainer interface {
	Application
	Get(string) interface{}
	Register(Provider)
	Bind(string, interface{})
	Alias(string, string)
	Singleton(string, func(app ApplicationContainer) interface{})
}

type ClazzInterface interface {
	fmt.Stringer
	Clazz() string
	Constructor() func() interface{}
	Factory() func(ApplicationContainer) interface{}
}

type RegisterInterface interface {
	Register()
}

type BootInterface interface {
	Boot()
}

type Provider interface {
	GetClazz() ClazzInterface
	Init(ApplicationContainer)
	GetSupportBean() SupportBean
	RegisterInterface
	BootInterface
}

type SupportBean struct {
	Register bool
	Boot     bool
}

type ConstructorInterface interface {
	Constructor() interface{}
}

type FactoryInterface interface {
	Factory(ApplicationContainer) interface{}
}
