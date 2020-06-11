package Components

import (
		"github.com/webGameLinux/kits/Contracts"
		"github.com/webGameLinux/kits/Libs"
		"os"
		"path/filepath"
		"strings"
		"sync"
)

type EnvironmentProvider interface {
		Contracts.Provider
		Set(key string, value string)
		Get(key string, defaults ...string) string
}

type EnvironmentComponents struct {
		FilePath string
		Storage  *HashMapperStrKeyEntry
}

type StrKeyEntry struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
}

type HashMapperStrKeyEntry struct {
		container []*StrKeyEntry
}

type HashIndex struct {
		Index, End int
		Exists     bool
}

type EnvironmentProviderImpl struct {
		manager *EnvironmentComponents
		bean    Contracts.SupportInterface
		clazz   Contracts.ClazzInterface
		app     Contracts.ApplicationContainer
		Name    string
}

type EnvironmentRegisterAfterFunc func(EnvironmentProvider)
type EnvironmentBootPrepareFunc func(Contracts.ApplicationContainer)
type EnvironmentFileLoaderFunc func(string) map[string]string

const (
		EnvironmentAlias                 = "env"
		EnvironmentLock                  = "env_lock"
		EnvFileDefault                   = ".env"
		EnvFileExt                       = ".env"
		EnvironmentFileLoader            = "EnvironmentFileLoader"
		EnvironmentProviderClass         = "EnvironmentProvider"
		EnvironmentProviderBootPrepare   = "EnvironmentProviderBootPrepare"
		EnvironmentProviderRegisterAfter = "EnvironmentProviderRegisterAfter"
)

func HashIndexOf(index, end int, exists bool) *HashIndex {
		var hashIndex = new(HashIndex)
		hashIndex.Index = index
		hashIndex.End = end
		hashIndex.Exists = exists
		return hashIndex
}

func HashMapperStrKeyEntryOf() *HashMapperStrKeyEntry {
		var (
				it []*StrKeyEntry
				m  = new(HashMapperStrKeyEntry)
		)
		m.container = it
		return m
}

func StrKeyEntryOf(args ...interface{}) *StrKeyEntry {
		var entry = new(StrKeyEntry)
		if len(args) >= 2 {
				entry.Key = args[0].(string)
				entry.Value = args[1]
		}
		return entry
}

var (
		environmentInstanceLock sync.Once
		environment             *EnvironmentProviderImpl
)

func environmentProviderNew() {
		environment = new(EnvironmentProviderImpl)
		environment.Name = EnvironmentProviderClass
}

func EnvironmentProviderOf() EnvironmentProvider {
		if environment == nil {
				environmentInstanceLock.Do(environmentProviderNew)
		}
		return environment
}

func EnvironmentComponentsOf(file ...string) *EnvironmentComponents {
		var component = new(EnvironmentComponents)
		if len(file) != 0 {
				component.FilePath = file[0]
		} else {
				component.FilePath = ""
		}
		component.Storage = HashMapperStrKeyEntryOf()
		return component
}

func (this *EnvironmentProviderImpl) GetClazz() Contracts.ClazzInterface {
		if this.clazz == nil {
				this.initClazz()
		}
		return this.clazz
}

func (this *EnvironmentProviderImpl) initClazz() {
		this.clazz = ClazzOf(this)
}

func (this *EnvironmentProviderImpl) Factory(container Contracts.ApplicationContainer) interface{} {
		this.Init(container)
		return this
}

func (this *EnvironmentProviderImpl) Constructor() interface{} {
		return EnvironmentProviderOf()
}

func (this *EnvironmentProviderImpl) Init(app Contracts.ApplicationContainer) {
		this.app = app
		this.init()
}

func (this *EnvironmentProviderImpl) init() {
		this.initClazz()
		this.initBean()
		this.initComponent()
}

func (this *EnvironmentProviderImpl) initComponent() {
		this.manager = EnvironmentComponentsOf()
}

func (this *EnvironmentProviderImpl) initBean() {
		this.bean = BeanOf()
}

func (this *EnvironmentProviderImpl) GetSupportBean() Contracts.SupportInterface {
		if this.bean == nil {
				this.initBean()
		}
		return this.bean
}

func (this *EnvironmentProviderImpl) Register() {
		// register env instance
		this.app.Bind(this.String(), this.manager)
		this.app.Bind(EnvironmentAlias, this)
		this.registerAfter()
}

func (this *EnvironmentProviderImpl) Boot() {
		// load env file
		this.loaderBootPrepare()
		this.loadEnvFile()
		// 监听 配置服务
}

func (this *EnvironmentProviderImpl) registerAfter() {
		afters := this.app.Get(EnvironmentProviderRegisterAfter)
		if afters == nil {
				return
		}
		if fn, ok := afters.(EnvironmentRegisterAfterFunc); ok {
				fn(this)
				return
		}
		if fn, ok := afters.(func(EnvironmentProvider)); ok {
				fn(this)
				return
		}
		if items, ok := afters.([]EnvironmentRegisterAfterFunc); ok {
				for _, fn := range items {
						fn(this)
				}
				return
		}
		if items, ok := afters.([]func(EnvironmentProvider)); ok {
				for _, fn := range items {
						fn(this)
				}
				return
		}
}

func (this *EnvironmentProviderImpl) loadEnvFile() {
		lock := this.app.Get(EnvironmentLock)
		if b, ok := lock.(bool); ok && b {
				return
		}
		loader := this.getEnvFileLoader()
		if loader == nil {
				return
		}
		file := this.getEnvFile()
		if file == "" {
				return
		}
		mapper := loader(file)
		if len(mapper) == 0 {
				return
		}
		for key, v := range mapper {
				this.Set(key, v)
		}
		this.app.Bind(EnvironmentLock, true)
}

func (this *EnvironmentProviderImpl) getEnvFile() string {
		basePath := this.app.GetProfile(Contracts.BasePath)
		if basePath == nil {
				basePath, _ = filepath.Abs(".")
		}
		if path, ok := basePath.(string); ok {
				mode := this.app.GetProfile(Contracts.RunModeEnv)
				if mode == nil {
						return path + string(filepath.Separator) + EnvFileExt
				}
				if m, ok := mode.(string); ok {
						return this.file(path, m)
				}
		}
		return EnvFileDefault
}

// env文件获取
func (this *EnvironmentProviderImpl) file(root string, mode string) string {
		var (
				file = root + string(filepath.Separator) + mode
		)
		switch mode {
		case Contracts.RunModeLocal:
				fallthrough
		case Contracts.RunModeStag:
				fallthrough
		case Contracts.RunModeTest:
				fallthrough
		case Contracts.RunModeDev:
				fallthrough
		case Contracts.RunModeProd:
		default:
				if state, err := os.Stat(file + EnvFileExt); err == nil {
						if !state.IsDir() {
								return file + EnvFileExt
						}
				}
				return EnvFileDefault
		}
		// mode env
		if state, err := os.Stat(file + EnvFileExt); err == nil {
				if !state.IsDir() {
						return file + EnvFileExt
				}
		}
		// mode dir env
		if state, err := os.Stat(file + string(filepath.Separator) + EnvFileExt); err == nil {
				if !state.IsDir() {
						return file + string(filepath.Separator) + EnvFileExt
				}
		}
		// .env
		return EnvFileDefault
}

func (this *EnvironmentProviderImpl) getEnvFileLoader() EnvironmentFileLoaderFunc {
		loader := this.app.Get(EnvironmentFileLoader)
		if loader == nil {
				return this.getEnvMapper
		}
		if fn, ok := loader.(EnvironmentFileLoaderFunc); ok {
				return fn
		}
		if fn, ok := loader.(func(string) map[string]string); ok {
				return fn
		}
		return this.getEnvMapper
}

func (this *EnvironmentProviderImpl) getEnvMapper(file string) map[string]string {
		var (
				mapper = make(map[string]string)
				loader = Libs.NewViperLoader()
		)
		loader.Mapper.SetConfigFile(file)
		loader.Mapper.SetConfigType(".env")
		if err := loader.Mapper.ReadInConfig(); err != nil {
				return mapper
		}
		loader.Foreach(func(k, v interface{}) bool {
				if key, ok := k.(string); ok {
						if value, ok := v.(string); ok {
								mapper[key] = value
						}
				}
				return true
		})
		return mapper
}

func (this *EnvironmentProviderImpl) loaderBootPrepare() {
		prepares := this.app.Get(EnvironmentProviderBootPrepare)
		if prepares == nil {
				return
		}
		if fn, ok := prepares.(EnvironmentBootPrepareFunc); ok {
				fn(this.app)
				return
		}
		if fn, ok := prepares.(func(Contracts.ApplicationContainer)); ok {
				fn(this.app)
				return
		}

		if items, ok := prepares.([]EnvironmentBootPrepareFunc); ok {
				for _, fn := range items {
						fn(this.app)
				}
				return
		}
		if items, ok := prepares.([]func(Contracts.ApplicationContainer)); ok {
				for _, fn := range items {
						fn(this.app)
				}
				return
		}
}

func (this *EnvironmentProviderImpl) Set(key string, value string) {
		this.manager.Storage.Set(key, value)
}

func (this *EnvironmentProviderImpl) Get(key string, defaults ...string) string {
		v := this.manager.Storage.GetStr(key, defaults...)
		if v == "" {
				v = os.Getenv(key)
				if v != "" {
						this.manager.Storage.Set(key, v)
				}
				return v
		}
		return ""
}

func (this *EnvironmentProviderImpl) String() string {
		return this.Name
}

// 获取
func (this *HashMapperStrKeyEntry) Get(key string) interface{} {
		keys := strings.SplitN(key, ".", -1)
		index, end, exists := this.find(keys)
		if index == -1 || !exists {
				return nil
		}
		hIndex := HashIndexOf(index, end, exists)
		return this.get(keys, hIndex)
}

// 获取值
func (this *HashMapperStrKeyEntry) get(keys []string, indexHash *HashIndex) interface{} {
		var (
				current interface{}
				index   = indexHash.Index
				end     = indexHash.End
				exists  = indexHash.Exists
		)
		if !exists {
				return nil
		}
		current = this.container[index]
		for i, key := range keys[1:] {
				if i >= end-1 {
						if entry, ok := current.(*StrKeyEntry); ok {
								if key != entry.Key {
										return nil
								}
								current = entry.Value
						}
						if i == end-1 {
								return current
						}
						if mapper, ok := current.(*HashMapperStrKeyEntry); ok {
								if i != end-1 {
										return mapper.get(keys[i:], HashIndexOf(1, end-i, exists))
								}
						}
				}
		}
		return nil
}

// 获取字符串
func (this *HashMapperStrKeyEntry) GetStr(key string, defaults ...string) string {
		v := this.Get(key)
		if len(defaults) == 0 {
				defaults = append(defaults, "")
		}
		if v == nil {
				return defaults[0]
		}
		if str, ok := v.(string); ok {
				return str
		}
		return defaults[0]
}

// 设置节点
func (this *HashMapperStrKeyEntry) Set(key string, value interface{}) {
		keys := strings.SplitN(key, ".", -1)
		index, end, exists := this.find(keys)
		if index == -1 && len(keys) == 1 {
				this.container = append(this.container, StrKeyEntryOf(keys[0], value))
				return
		}
		this.add(keys, value, HashIndexOf(index, end, exists))
}

// 查找
func (this *HashMapperStrKeyEntry) Search(search func(k, v, match interface{}) bool) interface{} {
		var res interface{}
		for _, it := range this.container {
				if !search(it.Key, it.Value, res) {
						break
				}
		}
		return res
}

// 节点数
func (this *HashMapperStrKeyEntry) Count() int {
		return len(this.container)
}

// 容量
func (this *HashMapperStrKeyEntry) Cap() int {
		return cap(this.container)
}

// 获取索引
func (this *HashMapperStrKeyEntry) Index(key string) int {
		if !strings.Contains(key, ".") {
				for i, entry := range this.container {
						if entry.Key == key {
								return i
						}
				}
				return -1
		}
		scopes := strings.SplitN(key, ".", -1)
		index, end, ok := this.find(scopes)
		if ok && end > 0 {
				return index
		}
		if index == -1 {
				return index
		}
		return -index
}

// 查找
func (this *HashMapperStrKeyEntry) find(scopes []string) (int, int, bool) {
		var (
				i         int
				key       string
				container interface{}
				index     int
				count     = len(scopes)
		)
		index = -1
		container = this.container
		for i, key = range scopes {
				switch container.(type) {
				case *HashMapperStrKeyEntry:
						if mapper, ok := container.(*HashMapperStrKeyEntry); ok {
								v := mapper.Index(key)
								if v == -1 {
										return -1, i, false
								}
								container = mapper.container[v].Value
								index = v
						}
				case *StrKeyEntry:
						if entry, ok := container.(*StrKeyEntry); ok {
								if entry.Key != key {
										return index, i, false
								}
								container = entry.Value
								if i >= count {
										i++
								}
						}
				default:
						return index, i, false
				}
		}
		return index, i, i > count
}

// 遍历
func (this *HashMapperStrKeyEntry) Foreach(each func(key, v interface{}) bool) {
		for _, entry := range this.container {
				if entry == nil {
						continue
				}
				if !each(entry.Key, entry.Value) {
						break
				}
		}
}

// 添加新节点
func (this *HashMapperStrKeyEntry) add(keys []string, value interface{}, indexHash *HashIndex) {
		var (
				current interface{}
				num     = len(keys) - 1
				index   = indexHash.Index
				end     = indexHash.End
				exists  = indexHash.Exists
		)
		current = this.container[index]
		for i, key := range keys[1:] {
				if end != num+2 {
						if i >= end-1 {
								if entry, ok := current.(*StrKeyEntry); ok {
										v := StrKeyEntryOf(key, nil)
										entry.Value = v
										if i != end-1 {
												it := HashMapperStrKeyEntryOf()
												v.Value = it
												current = it
										} else {
												v.Value = value
												return
										}
								}
								if mapper, ok := current.(*HashMapperStrKeyEntry); ok {
										if i != end-1 {
												it := HashMapperStrKeyEntryOf()
												mapper.Set(key, it)
												current = it
										} else {
												mapper.Set(key, value)
												return
										}
								}
						}
				}
				if entry, ok := current.(*StrKeyEntry); ok {
						if entry.Key == key {
								current = entry.Value
						}
						if exists && i+1 == num {
								entry.Value = value
						}
				}
				if mapper, ok := current.(*HashMapperStrKeyEntry); ok {
						if exists && i+1 != num {
								current = HashMapperStrKeyEntryOf()
								mapper.Set(key, current)
						}
						if exists && i+1 == num {
								mapper.Set(key, value)
								return
						}
				}
		}
}

// 过滤
func (this *HashMapperStrKeyEntry) Filter(filter func(key string, v interface{}) bool) *HashMapperStrKeyEntry {
		var mapper = HashMapperStrKeyEntryOf()
		for _, entry := range this.container {
				if entry == nil {
						continue
				}
				if filter(entry.Key, entry.Value) {
						mapper.Set(entry.Key, entry.Value)
				}
		}
		return mapper
}
