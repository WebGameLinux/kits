package Components

import (
		"fmt"
		"github.com/webGameLinux/kits/Contracts"
		"sync"
)

// 获取接口
type GetterInterface interface {
		Int(string, ...int) int
		Bool(string, ...bool) bool
		Get(string, ...string) string
		IntArray(string, ...[]int) []int
		FloatN(string, ...float64) float64
		Float(string, ...float32) float32
		Strings(string, ...[]string) []string
		Any(string, ...interface{}) interface{}
		Map(string, ...*map[string]interface{}) *map[string]interface{}
		HashMap(string, ...*HashMapperStrKeyEntry) *HashMapperStrKeyEntry
		Exists(string) bool
}

type SetterInterface interface {
		Set(string, interface{})
		Add(string, interface{})
		Remove(string)
}

// 配置服务
type ConfigureProvider interface {
		Contracts.Provider
		GetterInterface
}

type Configure struct {
		Mapper *sync.Map
}

type ConfigureProviderImpl struct {
		Name     string
		instance GetterInterface
		bean     *Contracts.SupportBean
		clazz    Contracts.ClazzInterface
		app      Contracts.ApplicationContainer
}

func (this *ConfigureProviderImpl) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.clazz = ClazzOf(this)
		}
		return this.clazz
}

func (this *ConfigureProviderImpl) Init(app Contracts.ApplicationContainer) {
		if this.app != nil {
				return
		}
		this.app = app
		this.instance = ConfigureOf()
		this.clazz = ClazzOf(this)
}

func (this *ConfigureProviderImpl) GetSupportBean() Contracts.SupportBean {
		if this.bean == nil {
				this.bean = BeanOf()
		}
		return *this.bean
}

func (this *ConfigureProviderImpl) Int(key string, defaults ...int) int {
		return this.instance.Int(key, defaults...)
}

func (this *ConfigureProviderImpl) Bool(key string, defaults ...bool) bool {
		return this.instance.Bool(key, defaults...)
}

func (this *ConfigureProviderImpl) Get(key string, defaults ...string) string {
		return this.instance.Get(key, defaults...)
}

func (this *ConfigureProviderImpl) IntArray(key string, defaults ...[]int) []int {
		return this.instance.IntArray(key, defaults...)
}

func (this *ConfigureProviderImpl) FloatN(key string, defaults ...float64) float64 {
		return this.instance.FloatN(key, defaults...)
}

func (this *ConfigureProviderImpl) Float(key string, defaults ...float32) float32 {
		return this.instance.Float(key, defaults...)
}

func (this *ConfigureProviderImpl) Strings(key string, defaults ...[]string) []string {
		return this.instance.Strings(key, defaults...)
}

func (this *ConfigureProviderImpl) Any(key string, defaults ...interface{}) interface{} {
		return this.instance.Any(key, defaults...)
}

func (this *ConfigureProviderImpl) Map(key string, defaults ...*map[string]interface{}) *map[string]interface{} {
		return this.instance.Map(key, defaults...)
}

func (this *ConfigureProviderImpl) HashMap(key string, defaults ...*HashMapperStrKeyEntry) *HashMapperStrKeyEntry {
		return this.instance.HashMap(key, defaults...)
}

func (this *ConfigureProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this
}

func (this *ConfigureProviderImpl) Constructor() interface{} {
		return ConfigureProviderOf()
}

func (this *ConfigureProviderImpl) String() string {
		return this.Name
}

func (this *ConfigureProviderImpl) Register() {
		this.app.Bind(this.String(), this)
		this.app.Bind("configure", this.instance)
}

func (this *ConfigureProviderImpl) Boot() {

}

func (this *ConfigureProviderImpl) Exists(key string) bool {
		return this.instance.Exists(key)
}

func ConfigureOf() GetterInterface {
		var configure = new(Configure)
		return configure
}

func ConfigureProviderOf() ConfigureProvider {
		var provider = new(ConfigureProviderImpl)
		provider.Name = "ConfigureProvider"
		return provider
}

func (this *Configure) Int(key string, defaults ...int) int {
		if len(defaults) == 0 {
				defaults = append(defaults, 0)
		}
		v := this.Any(key, defaults[0])
		// todo string to int
		if n, ok := v.(int); ok {
				return n
		}
		return defaults[0]
}

func (this *Configure) Bool(key string, defaults ...bool) bool {
		if len(defaults) == 0 {
				defaults = append(defaults, false)
		}
		v := this.Any(key, defaults[0])
		// todo string to bool
		if n, ok := v.(bool); ok {
				return n
		}
		return defaults[0]
}

func (this *Configure) Get(key string, defaults ...string) string {
		if len(defaults) == 0 {
				defaults = append(defaults, "")
		}
		v := this.Any(key, defaults[0])
		// todo string
		if n, ok := v.(string); ok {
				return n
		}
		if str, ok := v.(fmt.Stringer); ok {
				return str.String()
		}
		return defaults[0]
}

func (this *Configure) IntArray(key string, defaults ...[]int) []int {
		panic("implement me")
}

func (this *Configure) FloatN(key string, defaults ...float64) float64 {
		panic("implement me")
}

func (this *Configure) Float(key string, defaults ...float32) float32 {
		panic("implement me")
}

func (this *Configure) Strings(key string, defaults ...[]string) []string {
		panic("implement me")
}

func (this *Configure) Any(key string, defaults ...interface{}) interface{} {
		if len(defaults) == 0 {
				defaults = append(defaults, nil)
		}
		if v, ok := this.Mapper.Load(key); ok {
				return v
		}
		return defaults[0]
}

func (this *Configure) Map(key string, defaults ...*map[string]interface{}) *map[string]interface{} {
		panic("implement me")
}

func (this *Configure) HashMap(key string, defaults ...*HashMapperStrKeyEntry) *HashMapperStrKeyEntry {
		if len(defaults) == 0 {
				defaults = append(defaults, HashMapperStrKeyEntryOf())
		}
		v := this.Any(key, defaults[0])
		if h, ok := v.(*HashMapperStrKeyEntry); ok {
				return h
		}
		return nil
}

func (this *Configure) Set(key string, value interface{}) {
		if !this.Exists(key) {
				return
		}
		this.Mapper.Store(key, value)
}

func (this *Configure) Exists(key string) bool {
		if _, ok := this.Mapper.Load(key); ok {
				return true
		}
		return false
}

func (this *Configure) Add(key string, value interface{}) {
		this.Mapper.Store(key, value)
}

func (this *Configure) Remove(key string) {
		this.Mapper.Delete(key)
}
