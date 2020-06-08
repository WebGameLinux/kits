package Components

import (
		"fmt"
		"github.com/tietang/props/kvs"
		"github.com/webGameLinux/kits/Contracts"
		"io"
		"strings"
		"sync"
)

// 获取接口
type GetterInterface interface {
		Keys() []string
		Exists(string) bool
		Values() []interface{}
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
}

// 设置接口
type SetterInterface interface {
		Set(string, interface{})
		Add(string, interface{})
		Remove(string)
}

// 配置加载器
type ConfigureLoader func(config Configuration, app Contracts.ApplicationContainer)

// 配置
type Configuration interface {
		GetterInterface
		SetterInterface
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
		return this.config().Int(key, defaults...)
}

func (this *ConfigureProviderImpl) Bool(key string, defaults ...bool) bool {
		return this.config().Bool(key, defaults...)
}

func (this *ConfigureProviderImpl) Get(key string, defaults ...string) string {
		return this.config().Get(key, defaults...)
}

func (this *ConfigureProviderImpl) IntArray(key string, defaults ...[]int) []int {
		return this.config().IntArray(key, defaults...)
}

func (this *ConfigureProviderImpl) FloatN(key string, defaults ...float64) float64 {
		return this.config().FloatN(key, defaults...)
}

func (this *ConfigureProviderImpl) Float(key string, defaults ...float32) float32 {
		return this.config().Float(key, defaults...)
}

func (this *ConfigureProviderImpl) Strings(key string, defaults ...[]string) []string {
		return this.config().Strings(key, defaults...)
}

func (this *ConfigureProviderImpl) Any(key string, defaults ...interface{}) interface{} {
		return this.config().Any(key, defaults...)
}

func (this *ConfigureProviderImpl) Map(key string, defaults ...*map[string]interface{}) *map[string]interface{} {
		return this.config().Map(key, defaults...)
}

func (this *ConfigureProviderImpl) HashMap(key string, defaults ...*HashMapperStrKeyEntry) *HashMapperStrKeyEntry {
		return this.config().HashMap(key, defaults...)
}

func (this *ConfigureProviderImpl) Factory(app Contracts.ApplicationContainer) interface{} {
		this.Init(app)
		return this.instance
}

func (this *ConfigureProviderImpl) Keys() []string {
		return this.config().Keys()
}

func (this *ConfigureProviderImpl) Values() []interface{} {
		return this.config().Values()
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
		this.app.Alias("configure", "Configuration")
		this.app.Singleton("config", this.Factory)
		this.app.Bind("ConfigureLoader", ConfigLoader)
}

func (this *ConfigureProviderImpl) Boot() {
		configure := this.app.Get("Configuration")
		if cnf, ok := configure.(Configuration); ok {
				fn := this.app.Get("ConfigureLoader")
				if loader, ok := fn.(ConfigureLoader); ok {
						loader(cnf, this.app)
				}
		}
}

func (this *ConfigureProviderImpl) Exists(key string) bool {
		return this.instance.Exists(key)
}

func (this *ConfigureProviderImpl) config() Configuration {
		var configure = this.app.Get("config")
		if cnf, ok := configure.(Configuration); ok {
				return cnf
		}
		return nil
}

func ConfigureProviderOf() ConfigureProvider {
		var provider = new(ConfigureProviderImpl)
		provider.Name = "ConfigureProvider"
		return provider
}

func ConfigureOf() GetterInterface {
		var configure = new(Configure)
		return configure
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
		if n, ok := v.(bool); ok {
				return n
		}
		if v != nil {
				b := BooleanOf(v)
				if !b.Invalid() {
						return b.ValueOf()
				}
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
		if len(defaults) == 0 {
				defaults = append(defaults, []int{})
		}
		v := this.Any(key, defaults[0])
		if b, ok := v.([]int); ok {
				return b
		}
		if b, ok := v.(*[]int); ok {
				return *b
		}
		bv := IntArray(v)
		if !bv.Invalid() {
				return bv.ValueOf()
		}
		return defaults[0]
}

func (this *Configure) FloatN(key string, defaults ...float64) float64 {
		if len(defaults) == 0 {
				defaults = append(defaults, 0)
		}
		v := this.Any(key)
		if v == nil {
				return defaults[0]
		}
		if str, ok := v.(string); ok {
				if !IsNumber(str) {
						return defaults[0]
				}
				return NumberOf(str).FloatN()
		}
		if num, ok := v.(float64); ok {
				return num
		}
		return defaults[0]
}

func (this *Configure) Float(key string, defaults ...float32) float32 {
		if len(defaults) == 0 {
				defaults = append(defaults, 0)
		}
		v := this.Any(key)
		if v == nil {
				return defaults[0]
		}
		if str, ok := v.(string); ok {
				if !IsNumber(str) {
						return defaults[0]
				}
				return NumberOf(str).Float()
		}
		if num, ok := v.(float32); ok {
				return num
		}
		return defaults[0]
}

func (this *Configure) Strings(key string, defaults ...[]string) []string {
		if len(defaults) == 0 {
				defaults = append(defaults, []string{})
		}
		v := this.Any(key)
		if v == nil {
				return defaults[0]
		}
		if str, ok := v.(string); ok {
				if strings.Contains(str, ",") {
						return strings.SplitN(str, ",", -1)
				}
				return []string{str}
		}
		if strArr, ok := v.([]string); ok {
				return strArr
		}
		if strArr, ok := v.(*[]string); ok {
				return *strArr
		}
		arr := ArrayString(Array(v))
		if len(arr) != 0 {
				return arr
		}
		return defaults[0]
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
		if len(defaults) == 0 {
				defaults = append(defaults, new(map[string]interface{}))
		}
		v := this.Any(key, defaults[0])
		if h, ok := v.(*map[string]interface{}); ok {
				return h
		}
		return nil
}

func (this *Configure) HashMap(key string, defaults ...*HashMapperStrKeyEntry) *HashMapperStrKeyEntry {
		if len(defaults) == 0 {
				defaults = append(defaults, HashMapperStrKeyEntryOf())
		}
		v := this.Any(key, defaults[0])
		if h, ok := v.(*HashMapperStrKeyEntry); ok {
				return h
		}
		if h, ok := v.(*map[string]interface{}); ok {
				hash := HashMapperStrKeyEntryOf()
				for k, value := range *h {
						hash.Set(k, value)
				}
				return hash
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

func (this *Configure) Keys() []string {
		var keys []string
		this.Mapper.Range(func(key, value interface{}) bool {
				if str, ok := key.(string); ok {
						keys = append(keys, str)
				}
				return true
		})
		return keys
}

func (this *Configure) Values() []interface{} {
		var values []interface{}
		this.Mapper.Range(func(key, value interface{}) bool {
				values = append(values, value)
				return true
		})
		return values
}

// 获取配置读取方式
func ConfigLoader(config Configuration, app Contracts.ApplicationContainer) {
		// 文件读取器
		if files, ok := app.GetProfile("AppFile.Properties").([]string); ok {
				for _, fs := range files {
						scope := GetScope(fs)
						if prop, err := kvs.ReadProperties(GetFileReader(fs)); err == nil {
								for k, v := range prop.Values {
										config.Add(scope+k, v)
								}
						}
				}
		}
		// 读取器
		if reader, ok := app.GetProfile("AppFile.Properties.Reader").(io.Reader); ok {
				if prop, err := kvs.ReadProperties(reader); err == nil {
						for k, v := range prop.Values {
								config.Add(k, v)
						}
				}
		}
}

// 获取作用域
func GetScope(name string) string {
		if strings.Contains(name, "/") {
				strArr := strings.SplitN(name, "/", -1)
				name = strArr[len(strArr)-1]
		}
		if strings.Contains(name, `\`) {
				strArr := strings.SplitN(name, `\`, -1)
				name = strArr[len(strArr)-1]
		}
		if strings.Contains(name, ".") {
				strArr := strings.SplitN(name, ".", -1)
				num := len(strArr)
				if num >= 2 {
						name = strArr[num-2]
				} else {
						name = strArr[0]
				}
		}
		return name
}
