package WeChat

import (
		"encoding/json"
		"github.com/silenceper/wechat"
		"github.com/silenceper/wechat/cache"
		"os"
		"strings"
		"sync"
)

const (
		DefaultEntryKey          = "default"
		DefaultMemcacheUrl       = "127.0.0.1:11211"
		DefaultCacheRedis        = "redis"
		DefaultCacheMemCache     = "memcache"
		EnvWeChatPrefixKey       = "wechat"
		EnvWeChatCacheDriverKey  = "wechat_cache_driver"
		EnvWeChatAppIdKey        = "app_id"
		EnvWeChatAppSecretKey    = "app_secret"
		EnvWeChatEncodingAESKey  = "encoding_aes_key"
		EnvWeChatPayKey          = "pay_key"
		EnvWeChatPayMchIDKey     = "pay_mch_id"
		EnvWeChatTokenKey        = "token"
		EnvWeChatPayNotifyURLKey = "pay_notify_url"
		PrefixJoinDot            = "."
		EnvCacheEntryKey         = "cache"
		DriverSchemaDiv          = "@"
		DefaultRedisHost         = "127.0.0.1:6379"
)

var (
		instance     *ServiceImpl
		instanceLock sync.Once
)

type Service interface {
		SetCache(string, cache.Cache) Service
		Add(string, *wechat.Config) *wechat.Wechat
		Get(string, ...interface{}) *wechat.Wechat
		SetInstance(string, *wechat.Wechat) Service
}

type EntryConfig interface {
		Key() string
		Value() *wechat.Config
		Empty() bool
		Mapper() map[string]*wechat.Config
}

type EntryInstance interface {
		Key() string
		Empty() bool
		Value() *wechat.Wechat
}

type EntryConfigImpl struct {
		key   string
		value *wechat.Config
}

type EntryInstanceImpl struct {
		key   string
		value *wechat.Wechat
}

type ServiceImpl struct {
		mutex     sync.Mutex
		configure map[string]*wechat.Config
		instances map[string]*wechat.Wechat
		caches    map[string]cache.Cache
}

func NewWeChatService(args ...interface{}) *ServiceImpl {
		var service = new(ServiceImpl)
		service.init(args...)
		return service
}

func GetInstance() Service {
		if instance == nil {
				instanceLock.Do(func() {
						instance = NewWeChatService()
				})
		}
		return instance
}

func Get(name string, args ...interface{}) *wechat.Wechat {
		return GetInstance().Get(name, args...)
}

func Add(name string, ins *wechat.Config) *wechat.Wechat {
		return GetInstance().Add(name, ins)
}

func SetCache(key string, ins cache.Cache) Service {
		return GetInstance().SetCache(key, ins)
}

func SetConfig(key string, cnf *wechat.Config) Service {
		var ins = GetInstance()
		if impl, ok := ins.(*ServiceImpl); ok {
				impl.mutex.Lock()
				defer impl.mutex.Unlock()
				if _, ok := impl.configure[key]; ok {
						return ins
				}
				impl.configure[key] = cnf
		}
		return ins
}

func NewEntryWeChatInstance(key string, value *wechat.Wechat) *EntryInstanceImpl {
		return &EntryInstanceImpl{
				key: key, value: value,
		}
}

func NewEntryWeChatConfig(key string, value *wechat.Config) *EntryConfigImpl {
		return &EntryConfigImpl{
				key: key, value: value,
		}
}

//-------EntryInstanceImpl--------------------

func (this *EntryInstanceImpl) Key() string {
		return this.key
}

func (this *EntryInstanceImpl) Value() *wechat.Wechat {
		return this.value
}

func (this *EntryInstanceImpl) Empty() bool {
		return this.value == nil || this.key == ""
}

func (this *EntryInstanceImpl) Set(key string, value *wechat.Wechat) *EntryInstanceImpl {
		this.key = key
		this.value = value
		return this
}

//-------EntryConfig--------------------

func (this *EntryConfigImpl) Key() string {
		return this.key
}

func (this *EntryConfigImpl) Value() *wechat.Config {
		return this.value
}

func (this *EntryConfigImpl) Empty() bool {
		return this.key == "" || this.value == nil
}

func (this *EntryConfigImpl) Set(key string, value *wechat.Config) *EntryConfigImpl {
		this.key = key
		this.value = value
		return this
}

func (this *EntryConfigImpl) Mapper() map[string]*wechat.Config {
		return map[string]*wechat.Config{
				this.Key(): this.Value(),
		}
}

//-------Service---------------------

func (this *ServiceImpl) init(args ...interface{}) {
		if this.configure == nil {
				this.configure = make(map[string]*wechat.Config)
		}
		if this.instances == nil {
				this.instances = make(map[string]*wechat.Wechat)
		}
		if this.caches == nil {
				this.caches = make(map[string]cache.Cache)
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		for _, arg := range args {
				if m, ok := arg.(map[string]*wechat.Config); ok {
						for k, c := range m {
								this.configure[k] = c
						}
				}
				if entry, ok := arg.(EntryConfig); ok && !entry.Empty() {
						this.configure[entry.Key()] = entry.Value()
						continue
				}
				if entries, ok := arg.([]EntryConfig); ok {
						for _, en := range entries {
								if en.Empty() {
										continue
								}
								this.configure[en.Key()] = en.Value()
						}
						continue
				}
				if instance, ok := arg.(*wechat.Wechat); ok {
						this.instances[instance.Context.AppID] = instance
						continue
				}
				if instances, ok := arg.([]*wechat.Wechat); ok {
						for _, ins := range instances {
								if ins == nil {
										continue
								}
								this.instances[ins.Context.AppID] = ins
						}
						continue
				}
				if instance, ok := arg.(EntryInstance); ok && !instance.Empty() {
						this.instances[instance.Key()] = instance.Value()
						continue
				}
				if instances, ok := arg.([]EntryInstance); ok {
						for _, ins := range instances {
								if ins == nil {
										continue
								}
								this.instances[ins.Key()] = ins.Value()
						}
						continue
				}
		}
}

func (this *ServiceImpl) SetCache(key string, ca cache.Cache) Service {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if ca == nil {
				return this
		}
		this.caches[key] = ca
		return this
}

func (this *ServiceImpl) SetInstance(key string, ins *wechat.Wechat) Service {
		if ins, ok := this.instances[key]; ok && ins != nil {
				return this
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.instances[key] = ins
		return this
}

func (this *ServiceImpl) getConfig(entry ...string) *wechat.Config {
		if len(entry) == 0 {
				entry = append(entry, DefaultEntryKey)
		}
		if cnf, ok := this.configure[entry[0]]; ok && cnf != nil {
				return cnf
		}
		return this.getDefaults(entry...)
}

// 获取默认配置
func (this *ServiceImpl) getDefaults(prefixKey ...string) *wechat.Config {
		var (
				scope  string
				argc   = len(prefixKey)
				config = new(wechat.Config)
		)
		if argc <= 0 {
				prefixKey = append(prefixKey, EnvWeChatPrefixKey)
		}
		if prefixKey[0] != EnvWeChatPrefixKey {
				prefixKey = append([]string{EnvWeChatPrefixKey}, prefixKey...)
		}
		if len(prefixKey) < 2 {
				prefixKey = append(prefixKey, DefaultEntryKey)
		}
		// wechat.default.app_id
		scope = strings.Join(prefixKey, PrefixJoinDot)
		config.AppID = this.getConfigValueEnv(EnvWeChatAppIdKey, scope)
		config.AppSecret = this.getConfigValueEnv(EnvWeChatAppSecretKey, scope)
		config.EncodingAESKey = this.getConfigValueEnv(EnvWeChatEncodingAESKey, scope)
		config.PayKey = this.getConfigValueEnv(EnvWeChatPayKey, scope)
		config.PayMchID = this.getConfigValueEnv(EnvWeChatPayMchIDKey, scope)
		config.Token = this.getConfigValueEnv(EnvWeChatTokenKey, prefixKey...)
		config.PayNotifyURL = this.getConfigValueEnv(EnvWeChatPayNotifyURLKey, scope)
		config.Cache = this.getCache(scope)
		return config
}

// 获取缓存驱动
// @param string entry 缓存配置前缀｜缓存对象名｜缓存驱动@缓存配置前缀
func (this *ServiceImpl) getCache(entry ...string) cache.Cache {
		var (
				argc   = len(entry)
				driver string
		)
		if argc > 0 {
				ins := this.getCacheByName(entry[0])
				if ins != nil {
						return ins
				}
		} else {
				entry = append(entry, DefaultEntryKey)
		}
		driver = os.Getenv(EnvWeChatCacheDriverKey)
		if driver == "" {
				driver = DefaultCacheRedis
		}
		// 切换驱动
		if strings.Contains(entry[0], DriverSchemaDiv) {
				strArr := strings.SplitN(entry[0], DriverSchemaDiv, 2)
				if len(strArr) >= 2 {
						driver = strArr[0]
						entry[0] = strArr[1]
				}
		}
		key := entry[0]
		if !strings.Contains(key, PrefixJoinDot+EnvCacheEntryKey) {
				key = key + PrefixJoinDot + EnvCacheEntryKey
		}
		switch driver {
		case DefaultCacheMemCache:
				return cache.NewMemcache(this.getMemcacheUrl(key))
		case DefaultCacheRedis:
				return cache.NewRedis(this.getRedisOptions(key))
		}
		return nil
}

// 通过驱动器名获取缓存器
// @param string name 驱动器别名
func (this *ServiceImpl) getCacheByName(name string) cache.Cache {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if ins, ok := this.caches[name]; ok {
				return ins
		}
		return nil
}

// 通过环境变量获取配置
func (this *ServiceImpl) getConfigValueEnv(key string, prefix ...string) string {
		if len(prefix) > 0 {
				key = prefix[0] + PrefixJoinDot + key
		}
		return os.Getenv(key)
}

// 获取memcache url
func (this *ServiceImpl) getMemcacheUrl(name string) string {
		var url = os.Getenv(name)
		if url == "" {
				return DefaultMemcacheUrl
		}
		return url
}

// 获取redis配置
func (this *ServiceImpl) getRedisOptions(name string) *cache.RedisOpts {
		var (
				data    = os.Getenv(name)
				options = new(cache.RedisOpts)
		)
		if data == "" {
				options.Database = 0
				options.Host = DefaultRedisHost
				return options
		}
		_ = json.Unmarshal([]byte(data), options)
		return options
}

func (this *ServiceImpl) Get(name string, args ...interface{}) *wechat.Wechat {
		var (
				argc = len(args)
		)
		if argc > 0 {
				if cnf, ok := args[0].(*wechat.Config); ok {
						return this.Add(name, cnf)
				}
		}
		return this.getInstance(name)
}

func (this *ServiceImpl) Add(name string, cnf *wechat.Config) *wechat.Wechat {
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if ins, ok := this.instances[name]; ok {
				return ins
		}
		var ins = wechat.NewWechat(cnf)
		if _, ok := this.configure[name]; !ok {
				this.configure[name] = cnf
		}
		this.instances[name] = ins
		return ins
}

func (this *ServiceImpl) getInstance(name string) *wechat.Wechat {
		if ins, ok := this.instances[name]; ok && ins != nil {
				return ins
		}
		return this.make(name)
}

func (this *ServiceImpl) make(name string) *wechat.Wechat {
		var cnf = this.getConfig(name)
		if cnf.AppID == "" {
				return nil
		}
		return this.Add(name, cnf)
}
