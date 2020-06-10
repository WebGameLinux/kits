package Components

import (
		"fmt"
		"github.com/webGameLinux/kits/Contracts"
)

type SchemaServiceProvider interface {
		Contracts.Provider
		Schemas() []SchemaService
		Add(...SchemaService)
		fmt.Stringer
}

type SchemaService interface {
		InitializerInterface
		fmt.Stringer
		State() int
}

type SchemaServiceProviderImpl struct {
		Name        string
		bean        *Contracts.SupportBean
		clazz       Contracts.ClazzInterface
		SchemaLists []SchemaService
		app         Contracts.ApplicationContainer
		mapper      map[string]int
}

const (
		SchemaServiceProviderClass = "SchemaServiceProvider"
)

func SchemaServiceProviderOf(args ...interface{}) SchemaServiceProvider {
		var schema = new(SchemaServiceProviderImpl)
		schema.Name = SchemaServiceProviderClass
		schema.mapper = make(map[string]int)
		if len(args) != 0 {
				schema.init(args...)
		}
		return schema
}

func (this *SchemaServiceProviderImpl) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = ClazzOf(this)
		}
		return this.clazz
}

func (this *SchemaServiceProviderImpl) Init(app Contracts.ApplicationContainer) {
		if this.app == nil {
				this.app = app
		}
}

func (this *SchemaServiceProviderImpl) GetSupportBean() Contracts.SupportBean {
		if this.bean == nil {
				this.bean = BeanOf()
		}
		return *this.bean
}

func (this *SchemaServiceProviderImpl) Register() {
		if !this.app.Exists(this.String()) {
				this.app.Bind(this.String(), this)
		}
}

func (this *SchemaServiceProviderImpl) Boot() {
		for _, schema := range this.Schemas() {
				if schema.State() <= 0 {
						schema.Initializer(this.app)
						starter, ok := schema.(Contracts.Starter)
						if !ok {
								continue
						}
						if starter.Block() {
								go starter.StartUp()
						} else {
								starter.StartUp()
						}
				}
		}
}

func (this *SchemaServiceProviderImpl) Schemas() []SchemaService {
		provider := this.getProvider()
		if provider == this || provider == nil {
				return this.SchemaLists
		}
		return provider.Schemas()
}

func (this *SchemaServiceProviderImpl) Set(key string, v interface{}) {
		switch key {
		case "clazz":
				if class, ok := v.(Contracts.ClazzInterface); ok && this.clazz == nil {
						this.clazz = class
				}
		case "bean":
				if bean, ok := v.(Contracts.SupportBean); ok && this.bean == nil {
						this.bean = &bean
				}
				if bean, ok := v.(*Contracts.SupportBean); ok && this.bean == nil {
						this.bean = bean
				}
		case "class":
				if name, ok := v.(string); ok && this.Name == "" && name != "" {
						this.Name = name
				}
		case "schema_lists":
				if schemas, ok := v.([]SchemaService); ok {
						if this.SchemaLists == nil {
								this.SchemaLists = schemas
						} else {
								this.Add(schemas...)
						}
				}
		case "app":
				if app, ok := v.(Contracts.ApplicationContainer); ok && this.app == nil {
						this.app = app
				}
		}
}

func (this *SchemaServiceProviderImpl) Add(args ...SchemaService) {
		provider := this.getProvider()
		if provider == this || provider == nil {
				for _, service := range args {
						if !this.exists(service.String()) {
								this.SchemaLists = append(this.SchemaLists, service)
								this.add(service.String(), len(this.SchemaLists)-1)
						}
				}
				return
		}
		provider.Add(args...)
}

func (this *SchemaServiceProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}

func (this *SchemaServiceProviderImpl) Constructor() interface{} {
		return SchemaServiceProviderOf()
}

func (this *SchemaServiceProviderImpl) String() string {
		return this.Name
}

func (this *SchemaServiceProviderImpl) exists(id string) bool {
		if index, ok := this.mapper[id]; ok && index >= 0 {
				return true
		}
		return false
}

func (this *SchemaServiceProviderImpl) add(id string, index int) {
		this.mapper[id] = index
}

func (this *SchemaServiceProviderImpl) getProvider() SchemaServiceProvider {
		obj := this.app.Get(this.String())
		if provider, ok := obj.(SchemaServiceProvider); ok {
				return provider
		}
		return this
}

func (this *SchemaServiceProviderImpl) init(args ...interface{}) {
		for _, arg := range args {
				app, ok := arg.(Contracts.ApplicationContainer)
				if ok && this.app == nil {
						this.Init(app)
				}
				if provider, ok := arg.(SchemaServiceProvider); ok {
						if this.Name == "" {
								this.Name = provider.String()
						}
						if this.SchemaLists == nil {
								this.SchemaLists = provider.Schemas()
						} else {
								this.Add(provider.Schemas()...)
						}
				}
				if mapper, ok := arg.(map[string]interface{}); ok {
						this.initByMapper(mapper)
				}
				if mapper, ok := arg.(*map[string]interface{}); ok {
						this.initByMapper(*mapper)
				}
		}
}

func (this *SchemaServiceProviderImpl) Initializer(app ...Contracts.ApplicationContainer) {
		if len(app) == 0 {
				app = append(app, this.app)
		}
		if app[0] != nil {
				this.Init(app[0])
				this.Register()
		}
}

func (this *SchemaServiceProviderImpl) initByMapper(mapper map[string]interface{}) {
		for key, v := range mapper {
				this.Set(key, v)
		}
}
