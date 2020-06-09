package Libs

import (
		"github.com/spf13/viper"
		"os"
		"strings"
)

type SetterInterface interface {
		Set(string, interface{})
}

type SetterAnyInterface interface {
		Set(k, v interface{})
}

type AddInterface interface {
		Add(string, interface{})
}

type ConfigureViperLoader struct {
		Mapper *viper.Viper
		Read   int
		Remote bool
		watch  bool
}

// 参数
type ViperInitParam struct {
		Paths      []string               `json:"paths"`
		File       string                 `json:"file"`
		ConfigName string                 `json:"config_name"`
		ConfigType string                 `json:"config_type"`
		EnvPrefix  string                 `json:"env_prefix"`
		Defaults   map[string]interface{} `json:"defaults"`
		Remote     *ViperRemoteSource     `json:"remote"`
}

// 远程配置源设置
type ViperRemoteSource struct {
		Provider      string `json:"provider"`
		Endpoint      string `json:"endpoint"`
		Path          string `json:"path"`
		SecretKeyring string `json:"secret_keyring"`
}

var (
		paramKeys = []string{
				"paths",
				"file",
				"config_name",
				"config_type",
				"env_prefix",
				"defaults",
				"remote",
		}
		remoteSourceKeys = []string{"provider", "endpoint", "path", "secret_keyring"}
)

func ViperParamOf(args ...interface{}) *ViperInitParam {
		var param = new(ViperInitParam)
		if len(args) > 0 {
				param.init(args)
		}
		return param
}

func ViperRemoteSourceOf(args ...string) *ViperRemoteSource {
		var source = new(ViperRemoteSource)
		if len(args) > 0 {
				source.init(args...)
		}
		return source
}

// viper 配置加载器
func NewViperLoader(args ...interface{}) *ConfigureViperLoader {
		var loader = new(ConfigureViperLoader)
		loader.Mapper = viper.New()
		if len(args) != 0 {
				param := ViperParamOf(args...)
				if !param.IsEmpty() {
						loader.init(param)
				}
		}
		return loader
}

func (this *ConfigureViperLoader) init(params *ViperInitParam) {
		if params == nil || params.IsEmpty() {
				return
		}
		var err error
		this.initBase(params)
		this.Remote = this.initRemote(params)
		// 读取配置
		this.Read = -1
		if this.Remote {
				err = this.Mapper.ReadRemoteConfig()
		} else {
				err = this.Mapper.ReadInConfig()
		}
		if err == nil {
				this.Read = 1
		}
}

func (this *ConfigureViperLoader) initBase(params *ViperInitParam) {
		if len(params.Paths) != 0 {
				for _, path := range params.Paths {
						this.Mapper.AddConfigPath(path)
				}
		}
		if params.File != "" {
				this.Mapper.SetConfigFile(params.File)
		}
		if params.ConfigName != "" {
				this.Mapper.SetConfigName(params.ConfigName)
		}
		if params.ConfigType != "" {
				this.Mapper.SetConfigType(params.ConfigType)
		}
		if params.EnvPrefix != "" {
				this.Mapper.SetEnvPrefix(params.EnvPrefix)
		}
		if len(params.Defaults) != 0 {
				for k, v := range params.Defaults {
						this.Mapper.SetDefault(k, v)
				}
		}
}

func (this *ConfigureViperLoader) initRemote(params *ViperInitParam) bool {
		if params.Remote != nil && !params.Remote.IsEmpty() {
				source := params.Remote
				if source.SecretKeyring == "" && source.Provider != "" && source.Endpoint != "" {
						err := this.Mapper.AddRemoteProvider(source.Provider, source.Endpoint, source.Path)
						if err == nil {
								return true
						}
				}
				if source.SecretKeyring != "" && source.Provider != "" && source.Endpoint != "" {
						err := this.Mapper.AddSecureRemoteProvider(source.Provider, source.Endpoint, source.Path, source.SecretKeyring)
						if err == nil {
								return true
						}
				}
		}
		return false
}

func (this *ConfigureViperLoader) Get(key string) interface{} {
		return this.Mapper.Get(key)
}

func (this *ConfigureViperLoader) Keys() []string {
		return this.Mapper.AllKeys()
}

func (this *ConfigureViperLoader) Set(key string, v interface{}) *ConfigureViperLoader {
		this.Mapper.Set(key, v)
		return this
}

func (this *ConfigureViperLoader) Foreach(each func(k, v interface{}) bool) {
		keys := this.Keys()
		for _, key := range keys {
				if !each(key, this.Mapper.Get(key)) {
						break
				}
		}
}

func (this *ConfigureViperLoader) Search(search func(k, v, matches interface{}) bool) interface{} {
		var (
				matches interface{}
				keys    = this.Keys()
		)
		for _, key := range keys {
				if !search(key, this.Mapper.Get(key), matches) {
						break
				}
		}
		return matches
}

func (this *ConfigureViperLoader) CopyTo(v interface{}) {
		if this.IsEmpty() {
				return
		}
		if adder, ok := v.(AddInterface); ok {
				this.Foreach(func(k, v interface{}) bool {
						if key, ok := k.(string); ok && key != "" {
								adder.Add(key, v)
						}
						return true
				})
		}
		if setter, ok := v.(SetterInterface); ok {
				this.Foreach(func(k, v interface{}) bool {
						if key, ok := k.(string); ok && key != "" {
								setter.Set(key, v)
						}
						return true
				})
		}
		if setter, ok := v.(SetterAnyInterface); ok {
				this.Foreach(func(k, v interface{}) bool {
						setter.Set(k, v)
						return true
				})
		}
}

func (this *ConfigureViperLoader) Watch() {
		if !this.Remote {
				return
		}
		if this.watch {
				return
		}
		this.Mapper.WatchConfig()
		this.watch = true
}

func (this *ConfigureViperLoader) IsEmpty() bool {
		return len(this.Keys()) == 0
}

// 初始化
func (this *ViperInitParam) init(args ...interface{}) {
		for _, v := range args {
				if str, ok := v.(string); ok {
						if strings.Contains(str, ":") {
								this.kv(str)
						} else {
								this.initStr(str)
						}
				}
				if mapper, ok := v.(map[string]interface{}); ok && len(this.Defaults) == 0 {
						if this.mapperLike(mapper) {
								this.initByMap(mapper)
						} else {
								this.Defaults = mapper
						}
				}
				if paths, ok := v.([]string); ok && len(this.Paths) == 0 {
						this.Paths = paths
				}
				if param, ok := v.(*ViperInitParam); ok {
						this.merge(param)
				}
				if param, ok := v.(ViperInitParam); ok {
						this.merge(&param)
				}
				if remote, ok := v.(ViperRemoteSource); ok && this.Remote == nil {
						this.Remote = &remote
				}
				if remote, ok := v.(*ViperRemoteSource); ok && this.Remote == nil {
						this.Remote = remote
				}
		}
		if len(this.Paths) > 0 {
				this.Paths = unique(this.Paths)
		}
}

func (this *ViperInitParam) initByMap(mapper map[string]interface{}) {
		for _, key := range this.keys() {
				this.Set(key, mapper[key])
		}
}

func (this *ViperInitParam) Set(key string, v interface{}) {
		switch key {
		case "paths":
				paths, ok := v.([]string)
				if len(this.Paths) == 0 && ok && len(paths) > 0 {
						this.Paths = paths
				}
		case "file":
				file, ok := v.(string)
				if this.File == "" && ok && exists(file) {
						this.File = file
				}
		case "config_name":
				cnfName, ok := v.(string)
				if ok && this.ConfigName == "" && cnfName != "" {
						this.ConfigName = cnfName
				}
		case "config_type":
				cnfExt, ok := v.(string)
				if ok && this.ConfigType == "" && isSupportExt(cnfExt) {
						this.ConfigType = cnfExt
				}
		case "env_prefix":
				envPrefix, ok := v.(string)
				if ok && this.EnvPrefix == "" && envPrefix != "" {
						this.EnvPrefix = envPrefix
				}
		case "defaults":
				mapper, ok := v.(map[string]interface{})
				if ok && len(this.Defaults) == 0 && len(mapper) != 0 {
						this.Defaults = mapper
				}
		case "remote":
				mapper, ok := v.(map[string]string)
				if ok && this.Remote == nil && len(mapper) != 0 {
						this.Remote = ViperRemoteSourceOf()
						for k, v := range mapper {
								this.Remote.Set(k, v)
						}
				}
				if source, ok := v.(ViperRemoteSource); ok && this.Remote == nil {
						this.Remote = &source
						return
				}
				if source, ok := v.(*ViperRemoteSource); ok && this.Remote == nil && source != nil {
						this.Remote = source
				}
		}
}

func (this *ViperInitParam) mapperLike(mapper map[string]interface{}) bool {
		for _, key := range this.keys() {
				if _, ok := mapper[key]; !ok {
						return false
				}
		}
		return true
}

func (this *ViperInitParam) merge(param *ViperInitParam) {
		if len(this.Paths) == 0 && len(param.Paths) != 0 {
				this.Paths = param.Paths
		} else {
				this.Paths = append(this.Paths, param.Paths...)
		}
		if param.File != "" {
				this.File = param.File
		}
		if param.ConfigType != "" {
				this.ConfigType = param.ConfigType
		}
		if param.EnvPrefix != "" {
				this.EnvPrefix = param.EnvPrefix
		}
		if len(this.Defaults) != 0 {
				for k, v := range param.Defaults {
						this.Defaults[k] = v
				}
		}
}

func (this *ViperInitParam) kv(str string) {
		kvs := strings.SplitN(str, ":", 1)
		if len(kvs) != 2 || kvs[1] == "" {
				return
		}
		switch kvs[0] {
		case "file":
				if this.File == "" && exists(kvs[1]) {
						this.File = kvs[1]
				}
		case "config_name":
				if this.ConfigName == "" {
						this.ConfigName = kvs[1]
				}
		case "config_type":
				if this.ConfigType == "" && isSupportExt(kvs[1]) {
						this.ConfigType = kvs[1]
				}
		case "env_prefix":
				if this.EnvPrefix == "" {
						this.EnvPrefix = kvs[1]
				}
		}
}

func (this *ViperInitParam) initStr(str string) {
		if this.File == "" && exists(str) {
				this.File = str
		}

		if this.ConfigName == "" {
				this.ConfigName = str
		}

		if this.ConfigType == "" && isSupportExt(str) {
				this.ConfigType = str
		}

		if this.EnvPrefix == "" {
				this.EnvPrefix = str
		}
}

func (this *ViperInitParam) keys() []string {
		return paramKeys
}

// 是否为空
func (this *ViperInitParam) IsEmpty() bool {
		if len(this.Paths) != 0 {
				return false
		}
		if this.File != "" {
				return false
		}
		if this.ConfigName != "" {
				return false
		}
		if this.ConfigType != "" {
				return false
		}
		if this.EnvPrefix != "" {
				return false
		}
		if len(this.Defaults) != 0 {
				return false
		}
		return true
}

func (this *ViperRemoteSource) init(args ...string) {
		keys := this.keys()
		kArgc := len(keys)
		if kArgc == len(args) {
				for i, key := range this.keys() {
						this.Set(key, args[i])
				}
		} else {
				for i, v := range args {
						if strings.Contains(v, ":") {
								this.setKv(v)
								continue
						}
						if kArgc > i {
								this.Set(keys[i], v)
						}
				}
		}
}

func (this *ViperRemoteSource) setKv(v string) bool {
		kvs := strings.SplitN(v, ":", 1)
		if len(kvs) != 2 || kvs[1] == "" {
				return false
		}
		this.Set(kvs[0], kvs[1])
		return true
}

func (this *ViperRemoteSource) Set(key, value string) {
		switch key {
		case "path":
				this.Path = value
		case "endpoint":
				this.Endpoint = value
		case "provider":
				this.Provider = value
		case "secret_keyring":
				this.SecretKeyring = value
		}
}

func (this *ViperRemoteSource) Get(key string) string {
		switch key {
		case "path":
				return this.Path
		case "endpoint":
				return this.Endpoint
		case "provider":
				return this.Provider
		case "secret_keyring":
				return this.SecretKeyring
		}
		return ""
}

func (this *ViperRemoteSource) keys() []string {
		return remoteSourceKeys
}

func (this *ViperRemoteSource) IsEmpty() bool {
		if this.Path != "" {
				return false
		}
		if this.Endpoint != "" {
				return false
		}
		if this.Provider != "" {
				return false
		}
		if this.SecretKeyring != "" {
				return false
		}
		return true
}

// 去重
func unique(arr []string) []string {
		newArr := make([]string, 0)
		for i := 0; i < len(arr); i++ {
				repeat := false
				for j := i + 1; j < len(arr); j++ {
						if arr[i] == arr[j] {
								repeat = true
								break
						}
				}
				if !repeat {
						newArr = append(newArr, arr[i])
				}
		}
		return newArr
}

// 是否正确的支持类型
func isSupportExt(str string) bool {
		for _, ext := range viper.SupportedExts {
				if ext == str {
						return true
				}
		}
		return false
}

// 文件是否存在
func exists(fs string) bool {
		if _, err := os.Stat(fs); err == nil {
				return true
		}
		return false
}
