package Schemas

import (
		"github.com/astaxie/beego"
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Supports"
		"sync"
)

type BeegoHttpServerProvider interface {
		Contracts.Provider
		Server() *beego.App
		App() Contracts.ApplicationContainer
		Add(string, interface{})
}

type beegoHttpServerImpl struct {
		running   bool
		Name      string
		server    *beego.App
		clazz     Contracts.ClazzInterface
		bean      Contracts.SupportInterface
		app       Contracts.ApplicationContainer
		boots     []BootBeforeFn
		registers []RegisterBeforeFn
}

type BootBeforeFn func(BeegoHttpServerProvider)
type RegisterBeforeFn func(Contracts.ApplicationContainer)

var (
		instanceLock  sync.Once
		beegoInstance *beegoHttpServerImpl
)

const (
		Beego                = "Beego"
		BeegoHttpServerClass = "BeegoHttpServer"
		BeegoRegisterBefore  = "BeegoRegisterBefore"
		BeegoBootBefore      = "BeegoBootBefore"
		BootBeforeFnName     = "BootBeforeFnName"
		RegisterBeforeFnName = "RegisterBeforeFnName"
)

func newBeegoHttpServer() {
		beegoInstance = new(beegoHttpServerImpl)
		beegoInstance.defaults()
		beegoInstance.boots = []BootBeforeFn{}
		beegoInstance.registers = []RegisterBeforeFn{}
}

func BeegoHttpServerOf(args ...interface{}) BeegoHttpServerProvider {
		if beegoInstance == nil {
				instanceLock.Do(newBeegoHttpServer)
		}
		beegoInstance.init(args...)
		return beegoInstance
}

func (this *beegoHttpServerImpl) defaults() {
		this.Name = BeegoHttpServerClass
		if this.server == nil {
				this.server = beego.BeeApp
		}
}

func (this *beegoHttpServerImpl) Add(key string, fn interface{}) {
		if key == "" || fn == nil {
				return
		}
		switch key {
		case BootBeforeFnName:
				if f, ok := fn.(func(BeegoHttpServerProvider)); ok {
						this.boots = append(this.boots, f)
				}
		case RegisterBeforeFnName:
				if f, ok := fn.(func(Contracts.ApplicationContainer)); ok {
						this.registers = append(this.registers, f)
				}
		}
}

func (this *beegoHttpServerImpl) init(args ...interface{}) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				return
		}
		for _, arg := range args {
				if app, ok := arg.(beego.App); ok && this.server == nil {
						this.server = &app
						continue
				}
				if app, ok := arg.(*beego.App); ok && this.server == nil {
						this.server = app
						continue
				}
				if app, ok := arg.(Contracts.ApplicationContainer); ok && this.app == nil {
						this.app = app
						continue
				}
		}
}

func (this *beegoHttpServerImpl) Init(app Contracts.ApplicationContainer) {
		if this.app == nil {
				this.app = app
		}
}

func (this *beegoHttpServerImpl) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = Components.ClazzOf(this)
		}
		return this.clazz
}

func (this *beegoHttpServerImpl) GetSupportBean() Contracts.SupportInterface {
		if this.bean == nil {
				this.bean = Components.BeanOf()
		}
		return this.bean
}

func (this *beegoHttpServerImpl) Register() {
		this.register()
		if !this.app.Exists(Beego) {
				this.app.Bind(Beego, this.server)
		}
		if !this.app.Exists(this.String()) {
				this.app.Bind(this.String(), this)
		}
}

func (this *beegoHttpServerImpl) Boot() {
		if this.running == true {
				return
		}
		this.boot()
		go this.Server().Run()
}

func (this *beegoHttpServerImpl) App() Contracts.ApplicationContainer {
		if this.app == nil {
				this.app = Supports.App()
		}
		return this.app
}

func (this *beegoHttpServerImpl) register() {
		before := this.app.Get(BeegoRegisterBefore)
		for _, fn := range this.registers {
				fn(this.app)
		}
		if before == nil {
				return
		}
		if fn, ok := before.(RegisterBeforeFn); ok {
				fn(this.app)
				return
		}
		if fn, ok := before.(func(Contracts.ApplicationContainer)); ok {
				fn(this.app)
				return
		}
		if fnArr, ok := before.([]RegisterBeforeFn); ok {
				for _, fn := range fnArr {
						fn(this.app)
				}
				return
		}
		if fnArr, ok := before.([]func(Contracts.ApplicationContainer)); ok {
				for _, fn := range fnArr {
						fn(this.app)
				}
				return
		}
}

func (this *beegoHttpServerImpl) boot() {
		before := this.app.Get(BeegoBootBefore)
		for _, fn := range this.boots {
				fn(this)
		}
		if before == nil {
				return
		}
		if fn, ok := before.(BootBeforeFn); ok {
				fn(this)
				return
		}
		if fn, ok := before.(func(BeegoHttpServerProvider)); ok {
				fn(this)
				return
		}
		if fnArr, ok := before.([]BootBeforeFn); ok {
				for _, fn := range fnArr {
						fn(this)
				}
				return
		}
		if fnArr, ok := before.([]func(BeegoHttpServerProvider)); ok {
				for _, fn := range fnArr {
						fn(this)
				}
				return
		}
}

func (this *beegoHttpServerImpl) String() string {
		return this.Name
}

func (this *beegoHttpServerImpl) Server() *beego.App {
		if this.app == nil {
				this.app = Supports.App()
		}
		server := this.app.Get(Beego)
		if server == nil {
				return this.server
		}
		if app, ok := server.(*beego.App); ok {
				return app
		}
		return this.server
}

func (this *beegoHttpServerImpl) Constructor() interface{} {
		return BeegoHttpServerOf()
}

func (this *beegoHttpServerImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}
