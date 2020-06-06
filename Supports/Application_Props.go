package Supports

import (
	"github.com/webGameLinux/kits/Components"
	"github.com/webGameLinux/kits/Contracts"
	"reflect"
)

type ApplicationProps struct {
	Providers         []Contracts.Provider
	ConfigFilesSuffix []string
	AppName           string
	Version           string
	ApkPath           string
	CtrChan           chan bool
}

var defaultProps *ApplicationProps

// 常量
const (
	appName        = "app"
	appVersionName = "1.0.0"
)

// 获取应用默认属性配置 | 内置配置
func getApplicationDefaultProps() *ApplicationProps {
	if defaultProps == nil {
		getSafeLock(defaultPropsLock).Do(defaultPropsFactory)
	}
	return defaultProps
}

func defaultPropsFactory() {
	defaultProps = new(ApplicationProps)
	defaultProps.init()
}

func (this *ApplicationProps) init() *ApplicationProps {
	this.initKeyValues()
	this.initProviders()
	return this
}

func (this *ApplicationProps) GetProviders() []Contracts.Provider {
	return this.Providers
}

func (this *ApplicationProps) initKeyValues() {
	this.AppName = appName
	this.Version = appVersionName
	this.CtrChan = make(chan bool, 2)
	this.ApkPath = reflect.TypeOf(this).PkgPath()
	this.ConfigFilesSuffix = []string{".yml", ".properties", ".ini"}
}

func (this *ApplicationProps) initProviders() {
	this.Providers = []Contracts.Provider{
		Components.CommandLineArgsProviderOf(), // commandLine
		Components.EnvironmentProviderOf(),     // environment
	}
}

// 获取
func (this *ApplicationProps) Get(key string) interface{} {
	switch key {
	case "Providers":
		fallthrough
	case "providers":
		return this.Providers
	case "AppName":
		fallthrough
	case "appname":
		fallthrough
	case "app_name":
		return this.AppName
	case "Version":
	case "version":
		return this.Version
	case "ApkPath":
		fallthrough
	case "apk_path":
		fallthrough
	case "apkpath":
		return this.ApkPath
	case "ConfigFilesSuffix":
		fallthrough
	case "configfilessuffix":
		fallthrough
	case "config_files_suffix":
		return this.ConfigFilesSuffix
	case "CtrChan":
		fallthrough
	case "ctrchan":
		fallthrough
	case "ctr_chan":
		return this.CtrChan

	}
	return nil
}

// 遍历
func (this *ApplicationProps) Foreach(each func(key string, value interface{}) bool) {
	for _, key := range this.keys() {
		if !each(key, this.Get(key)) {
			break
		}
	}
}

// 所有keys
func (this *ApplicationProps) keys() []string {
	return []string{
		"Providers", "AppName", "Version",
		"ApkPath", "ConfigFilesSuffix","CtrChan",
	}
}
