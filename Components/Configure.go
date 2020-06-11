package Components

import (
		"fmt"
		"github.com/tietang/props/kvs"
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Libs"
		"io"
		"os"
		"path/filepath"
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
		Foreach(func(k, v interface{}) bool)
		Search(func(k, v, match interface{}) bool) interface{}
		Load(interface{})
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
		Inject(scope string, obj interface{}, tags ...string) bool
}

type Configure struct {
		Mapper *sync.Map
}

type ConfigureProviderImpl struct {
		Name     string
		instance GetterInterface
		bean     Contracts.SupportInterface
		clazz    Contracts.ClazzInterface
		app      Contracts.ApplicationContainer
}

const (
		ConfigAlias            = "config"
		ConfigureAlias         = "configure"
		ConfigurationAlias     = "Configuration"
		ConfigureLoaderName    = "ConfigureLoader"
		ConfigureProviderClass = "ConfigureProvider"
)

var (
		configureInstanceLock sync.Once
		configureProvider     *ConfigureProviderImpl
)

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

func (this *ConfigureProviderImpl) GetSupportBean() Contracts.SupportInterface {
		if this.bean == nil {
				this.bean = BeanOf()
		}
		return this.bean
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

func (this *ConfigureProviderImpl) Foreach(each func(k, v interface{}) bool) {
		this.config().Foreach(each)
}

func (this *ConfigureProviderImpl) Search(search func(k, v, matches interface{}) bool) interface{} {
		return this.config().Search(search)
}

// 配置注入
// scope 作用域|前缀
// obj   struct
// tag  struct 注入tag, 1-3个支持,默认使用json tag 注入
func (this *ConfigureProviderImpl) Inject(scope string, obj interface{}, tags ...string) bool {
		var (
				injector = NewInjector(tags...)
				tagArr   = injector.Keys(obj)
		)
		if tagArr == nil || len(tagArr) == 0 {
				return false
		}
		mapper, tag := this.getScopes(scope, tagArr)
		if len(mapper) == 0 || tag == "" {
				return false
		}
		return injector.Copy(mapper, obj, tag)
}

// 批量获取
// 值 mapper, tag
// tag == "" 表示不完整获取
func (this *ConfigureProviderImpl) getScopes(root string, keys interface{}) (map[string]interface{}, string) {
		var (
				n      int
				tag    = "default"
				mapper = make(map[string]interface{})
		)
		if keys == nil {
				return mapper, ""
		}
		// 批量获取的
		if arr, ok := keys.([]string); ok {
				mapper, n = this.getScopesByArr(root, arr)
				if n == len(arr) && len(mapper) > 0 {
						return mapper, tag
				}
				return mapper, ""
		}
		// 分组获取的
		if mapArr, ok := keys.(map[string][]string); ok {
				var arr []string
				for tag, arr = range mapArr {
						mapper, n = this.getScopesByArr(root, arr)
						if n == len(arr) && len(mapper) > 0 {
								return mapper, tag
						}
				}
		}
		return mapper, ""
}

// 批量获取
// 值 ，长度
func (this *ConfigureProviderImpl) getScopesByArr(root string, keys []string) (map[string]interface{}, int) {
		if len(keys) == 0 {
				return map[string]interface{}{}, 0
		}
		var mapper = make(map[string]interface{})
		for _, key := range keys {
				k := key
				if root != "" {
						key = root + "." + key
				}
				val := this.Any(key)
				if val == nil {
						continue
				}
				mapper[k] = val
		}
		return mapper, len(mapper)
}

func (this *ConfigureProviderImpl) Constructor() interface{} {
		return ConfigureProviderOf()
}

func (this *ConfigureProviderImpl) String() string {
		return this.Name
}

func (this *ConfigureProviderImpl) Load(v interface{}) {
		this.config().Load(v)
}

func (this *ConfigureProviderImpl) Register() {
		this.app.Bind(this.String(), this)
		this.app.Bind(ConfigureAlias, this.instance)
		this.app.Alias(ConfigureAlias, ConfigurationAlias)
		this.app.Singleton(ConfigAlias, this.Factory)
		this.app.Bind(ConfigureLoaderName, ConfigLoader)
}

func (this *ConfigureProviderImpl) Boot() {
		configure := this.app.Get(ConfigurationAlias)
		if cnf, ok := configure.(Configuration); ok {
				fn := this.app.Get(ConfigureLoaderName)
				if loader, ok := fn.(ConfigureLoader); ok {
						loader(cnf, this.app)
				}
				if loader, ok := fn.(func(Configuration, Contracts.ApplicationContainer)); ok {
						loader(cnf, this.app)
				}
		}
}

func (this *ConfigureProviderImpl) Exists(key string) bool {
		return this.instance.Exists(key)
}

func (this *ConfigureProviderImpl) config() Configuration {
		var configure = this.app.Get(ConfigAlias)
		if cnf, ok := configure.(Configuration); ok {
				return cnf
		}
		return nil
}

func ConfigureProviderOf() ConfigureProvider {
		if configureProvider == nil {
				configureInstanceLock.Do(configureProviderNew)
		}
		return configureProvider
}

func configureProviderNew() {
		configureProvider = new(ConfigureProviderImpl)
		configureProvider.Name = ConfigureProviderClass
}

func ConfigureOf() Configuration {
		var configure = new(Configure)
		configure.Mapper = &sync.Map{}
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

// 遍历接口
func (this *Configure) Foreach(each func(k, v interface{}) bool) {
		this.Mapper.Range(each)
}

// 查找接口
func (this *Configure) Search(search func(k, v, match interface{}) bool) interface{} {
		var matches interface{}
		this.Mapper.Range(func(key, value interface{}) bool {
				if !search(key, value, matches) {
						return false
				}
				return true
		})
		return matches
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

func (this *Configure) Loader(params ConfigLoaderParams) {
		if params != nil {
				params.Load(this)
		}
}

// 添加载入｜配置
func (this *Configure) Load(v interface{}) {
		if params, ok := v.(ConfigLoaderParams); ok && !params.IsEmpty() {
				this.Loader(params)
		}
		params := ConfigLoaderParamsOf(v)
		if !params.IsEmpty() {
				this.Loader(params)
		}
}

type ConfigLoaderParamsImpl struct {
		Object interface{}                   // []string| configure | map |reader
		Loader func(configure Configuration) // 自定义加载器
}

type ConfigLoaderParams interface {
		Load(configuration Configuration)
		IsEmpty() bool
}

// 参数加载器
func ConfigLoaderParamsOf(args ...interface{}) ConfigLoaderParams {
		var params = new(ConfigLoaderParamsImpl)
		if len(args) > 0 {
				for _, v := range args {
						if str, ok := v.(string); ok && params.Object == nil {
								files := MakeFiles(str)
								if len(files) == 0 {
										continue
								}
								params.Object = files
								continue
						}
						if arr, ok := v.([]string); ok && params.Object == nil {
								params.Object = arr
								continue
						}
						if config, ok := v.(Configuration); ok && params.Object == nil {
								params.Object = config
								continue
						}
						if reader, ok := v.(io.Reader); ok && params.Object == nil {
								params.Object = reader
								continue
						}
						if loader, ok := v.(func(configure Configuration)); ok && params.Loader == nil {
								params.Loader = loader
								continue
						}
						// 类型 mapper
						if mapper, ok := v.(map[string]interface{}); ok && params.Object == nil {
								params.Object = mapper
						}
						if mapper, ok := v.(*map[string]interface{}); ok && params.Object == nil {
								params.Object = *mapper
						}
						if obj, ok := v.(ConfigLoaderParams); ok && !obj.IsEmpty() {
								if params.Object == nil && params.Loader == nil {
										return obj
								}
						}
				}
		}
		return params
}

func (this *ConfigLoaderParamsImpl) IsEmpty() bool {
		if this.Object == nil && this.Loader == nil {
				return true
		}
		return false
}

func (this *ConfigLoaderParamsImpl) Load(Cnf Configuration) {
		// 加载器加载
		if this.Loader != nil {
				this.Loader(Cnf)
		}
		// 对象加载器
		if this.Object != nil {
				if arr, ok := this.Object.([]string); ok {
						for _, fs := range arr {
								scope := GetScope(fs)
								// 非空文件
								if IsFile(fs) != 1 {
										continue
								}
								if prop, err := kvs.ReadProperties(GetFileReader(fs)); err == nil {
										for k, v := range prop.Values {
												Cnf.Add(scope+"."+k, v)
										}
								}
						}
				}
				if config, ok := this.Object.(Configuration); ok {
						config.Foreach(func(k, v interface{}) bool {
								if str, ok := k.(string); ok {
										Cnf.Add(str, v)
										return true
								}
								if str, ok := k.(fmt.Stringer); ok {
										Cnf.Add(str.String(), v)
								}
								return true
						})
				}
				// reader
				if reader, ok := this.Object.(io.Reader); ok {
						if props, err := kvs.ReadProperties(reader); err == nil {
								for k, v := range props.Values {
										Cnf.Add(k, v)
								}
						}
				}
				// 类型 mapper
				if mapper, ok := this.Object.(map[string]interface{}); ok {
						for k, v := range mapper {
								Cnf.Add(k, v)
						}
				}
		}
}

// 获取配置读取方式
func ConfigLoader(config Configuration, app Contracts.ApplicationContainer) {
		// 文件读取器
		if files, ok := app.GetProfile("App.Properties.files").([]string); ok {
				for _, fs := range files {
						scope := GetScope(fs)
						if prop, err := kvs.ReadProperties(GetFileReader(fs)); err == nil {
								for k, v := range prop.Values {
										config.Add(scope+"."+k, v)
								}
						}
				}
		}
		// 读取器
		if reader, ok := app.GetProfile("App.Properties.Reader").(io.Reader); ok {
				if prop, err := kvs.ReadProperties(reader); err == nil {
						for k, v := range prop.Values {
								config.Add(k, v)
						}
				}
		}
}

// 毒蛇加载器 读取配置
func ViperConfigLoader(config Configuration, app Contracts.ApplicationContainer) {
		// 文件读取器
		if paths, ok := app.GetProfile("App.Properties.Paths").([]string); ok {
				loader := Libs.NewViperLoader()
				for _, path := range paths {
						state, err := os.Stat(path)
						if err != nil || !state.IsDir() {
								continue
						}
						_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
								if !info.IsDir() {
										if strings.Contains(path, info.Name()) {
												_ = getFileConfigure(config, loader, "", path)
										} else {
												_ = getFileConfigure(config, loader, path, info.Name())
										}
								}
								return nil
						})
				}

		}
		if files, ok := app.GetProfile("App.Properties.files").([]string); ok {
				for _, file := range files {
						jsonArg := strings.Contains(file, "{") && strings.Contains(file, "}")
						if jsonArg {
								loader := Libs.NewViperLoader(file)
								loader.CopyTo(config)
								continue
						}
						// tagArg eg: paths:[];;;
						tagArg := strings.Contains(file, ":") && strings.Count(file, ":") > 2
						if jsonArg {
								loader := Libs.NewViperLoader(tagArg)
								loader.CopyTo(config)
								continue
						}
						loader := Libs.NewViperLoader()
						if file == "." || file == ".." {
								continue
						}
						if !filepath.IsAbs(file) {
								root := app.GetProfile("BasePath")
								file = strings.Replace(file, "./", "", 1)
								if r, ok := root.(string); ok && filepath.IsAbs(r) {
										file = r + string(filepath.Separator) + file
								} else {
										abs, _ := filepath.Abs(".")
										file = abs + string(filepath.Separator) + file
								}
						}

						_ = getFileConfigure(config, loader, "", file)

				}
		}
		// 读取器
		if reader, ok := app.GetProfile("App.Properties.Reader").(io.Reader); ok {
				loader := Libs.NewViperLoader()
				err := loader.Mapper.ReadConfig(reader)
				if err == nil {
						loader.CopyTo(config)
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

func getFileConfigure(config interface{}, loader *Libs.ConfigureViperLoader, path string, filename string) error {
		var (
				fs  = filename
				ext = ""
		)
		if path != "" {
				fs = path + string(filepath.Separator) + filename
		}
		LoggerProviderOf().Info("config loader file :" + fs)
		if strings.Contains(fs, string(filepath.Separator)) {
				arr := strings.SplitN(fs, string(filepath.Separator), -1)
				ext = strings.Replace(filepath.Ext(arr[len(arr)-1]), ".", "", -1)
		} else {
				ext = strings.Replace(filepath.Ext(fs), ".", "", -1)
		}
		loader.Mapper.SetConfigFile(fs)
		loader.Mapper.SetConfigType(ext)
		if err := loader.Mapper.ReadInConfig(); err != nil {
				return err
		}
		loader.CopyTo(config)
		return nil
}
