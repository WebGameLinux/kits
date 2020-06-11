package Components

import (
		"github.com/webGameLinux/kits/Contracts"
		"sync"
)

type ServiceProvider interface {
		Contracts.Provider
}

type AppServiceProvider struct {
		Name  string
		Clazz Contracts.ClazzInterface
		Bean  Contracts.SupportInterface
		app   Contracts.ApplicationContainer
}

var (
		serviceProviderLock sync.Once
		serviceProvider *AppServiceProvider
)

const (
		ServiceProviderClass = "ServiceProvider"
)

func ServiceProviderOf() ServiceProvider  {
		if serviceProvider == nil {
				serviceProviderLock.Do(newServiceProvider)
		}
		return serviceProvider
}

func newServiceProvider()  {
		serviceProvider = new(AppServiceProvider)
		serviceProvider.Name = ServiceProviderClass
}

func (this *AppServiceProvider)init()  {
		this.initBean()
		this.initClazz()
}

func (this *AppServiceProvider)initBean()  {
		if this.Bean == nil {
				this.Bean = BeanOf()
		}
}

func (this *AppServiceProvider)initClazz()  {
		if this.Clazz == nil {
				this.Clazz = ClazzOf(this)
		}
}

func (this *AppServiceProvider)Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}

func (this *AppServiceProvider)Constructor() interface{}  {
		return ServiceProviderOf()
}

func (this *AppServiceProvider)Init(app Contracts.ApplicationContainer)  {
		if this.app == nil {
				this.app = app
		}
}

func (this *AppServiceProvider) GetClazz() Contracts.ClazzInterface {
		if this.Clazz == nil {
				this.initClazz()
		}
		return this.Clazz
}

func (this *AppServiceProvider) GetSupportBean() Contracts.SupportInterface {
		if this.Bean == nil {
				this.initBean()
		}
		return this.Bean
}

func (this *AppServiceProvider)String()string  {
		return this.Name
}

func (this *AppServiceProvider) Register() {

}

func (this *AppServiceProvider) Boot() {

}

