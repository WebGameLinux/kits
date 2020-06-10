package Components

import (
		"fmt"
		"github.com/webGameLinux/kits/Contracts"
		"sync"
)

type InitializerInterface interface {
		Initializer(...Contracts.ApplicationContainer)
}

type Bootstrapper interface {
		fmt.Stringer
		InitializerInterface
		Contracts.BootInterface
}

type BootstrapProvider interface {
		Contracts.Provider
		Booted() bool
		Remove(string) bool
		Add(...Bootstrapper) bool
		GetBoots() []string
		GetBootstrapper() map[string]Bootstrapper
		LoadBootstrapper(string, ...bool) bool
		InitializerInterface
}

type appBootstrapper struct {
		Name   string
		booted bool
		Lists  map[string]Bootstrapper
		app    Contracts.ApplicationContainer
		clazz  Contracts.ClazzInterface
		bean   *Contracts.SupportBean
}

var (
		instanceLock sync.Once
		bootstrapper *appBootstrapper
)

const (
		AppBootstrapperClass = "AppBootstrapper"
)

func bootstrapperNew() {
		bootstrapper = new(appBootstrapper)
		bootstrapper.Name = AppBootstrapperClass
		bootstrapper.Lists = make(map[string]Bootstrapper)
}

// Bootstrapper
func AppBootstrapperOf(args ...Bootstrapper) BootstrapProvider {
		if bootstrapper == nil {
				instanceLock.Do(bootstrapperNew)
		}
		if len(args) != 0 && !bootstrapper.Booted() {
				bootstrapper.Add(args...)
		}
		return bootstrapper
}

func (this *appBootstrapper) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = ClazzOf(this)
		}
		return this.clazz
}

func (this *appBootstrapper) Init(app Contracts.ApplicationContainer) {
		if this.app == nil {
				this.app = app
		}
}

func (this *appBootstrapper) GetSupportBean() Contracts.SupportBean {
		if this.bean == nil {
				this.bean = BeanOf()
		}
		return *this.bean
}

func (this *appBootstrapper) Register() {
		if !this.app.Exists(this.String()) {
				this.app.Bind(this.String(), this)
		}
		// 初始化
		for _, boot := range this.GetBootstrapper() {
				boot.Initializer(this.app)
		}
}

func (this *appBootstrapper) Boot() {
		this.Initializer(this.app)
}

func (this *appBootstrapper) Booted() bool {
		return this.booted
}

func (this *appBootstrapper) GetBootstrapper() map[string]Bootstrapper {
		var provider = this.getMasterBootstrapper()
		if provider == this || provider == nil {
				return this.Lists
		}
		return provider.GetBootstrapper()
}

func (this *appBootstrapper) getMasterBootstrapper() BootstrapProvider {
		boot := this.app.Get(this.String())
		if provider, ok := boot.(BootstrapProvider); ok {
				return provider
		}
		return this
}

func (this *appBootstrapper) Initializer(app ...Contracts.ApplicationContainer) {
		if this.Booted() {
				return
		}
		if len(app) == 0 {
				app = append(app, this.app)
		}
		for _, boot := range this.GetBootstrapper() {
				boot.Boot()
		}
		this.booted = true
}

func (this *appBootstrapper) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}

func (this *appBootstrapper) Constructor() interface{} {
		return AppBootstrapperOf()
}

func (this *appBootstrapper) Add(items ...Bootstrapper) bool {
		var ok bool
		for _, it := range items {
				if it == nil {
						continue
				}
				key := it.String()
				if obj, ok := this.Lists[key]; ok && obj != nil {
						continue
				}
				this.Lists[key] = it
				ok = true
		}
		return ok
}

func (this *appBootstrapper) Remove(key string) bool {
		if _, ok := this.Lists[key]; ok {
				delete(this.Lists, key)
				return true
		}
		return false
}

func (this *appBootstrapper) GetBoots() []string {
		var boots []string
		for key, obj := range this.GetBootstrapper() {
				if key == obj.String() {
						boots = append(boots, key)
				}
		}
		return boots
}

func (this *appBootstrapper) LoadBootstrapper(key string, force ...bool) bool {
		if len(force) == 0 {
				force = append(force, false)
		}
		if this.Booted() {
				if !force[0] {
						return false
				}
		}
		if obj, ok := this.Lists[key]; ok && obj.String() == key {
				obj.Initializer(this.app)
				obj.Boot()
				return true
		}
		return false
}

func (this *appBootstrapper) String() string {
		return this.Name
}
