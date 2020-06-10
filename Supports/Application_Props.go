package Supports

import (
		"github.com/webGameLinux/kits/Components"
		"github.com/webGameLinux/kits/Contracts"
		"os"
		"path/filepath"
		"reflect"
)

type ApplicationProps struct {
		Providers         []Contracts.Provider
		ConfigFilesSuffix []string
		AppName           string
		Version           string
		ApkPath           string
		BasePath          string
		ConfigDir         string
		RunMode           string
		CtrChan           chan int
}

// 常量
const (
		appName        = "app"
		appVersionName = "1.0.0"
		RunModeDev     = "dev"
		RunModeTest    = "test"
		RunModeLocal   = "local"
		RunModelProd   = "prod"
		RunModelStag   = "stg"
		RunModelEnv    = "RunMode"
		ConfigDirEnv   = "ConfigDir"
)

var (
		defaultProps    *ApplicationProps
		supportRunModes = []string{
				RunModeDev,
				RunModeTest,
				RunModeLocal,
				RunModelProd,
				RunModelStag,
		}
)

// 获取应用默认属性配置 | 内置配置
func getApplicationDefaultProps() *ApplicationProps {
		if defaultProps == nil {
				getSafeLock(defaultPropsLock).Do(defaultPropsFactory)
		}
		return defaultProps
}

// 获取应用默认配置
func ApplicationDefaultProps() *ApplicationProps {
		return getApplicationDefaultProps()
}

func defaultPropsFactory() {
		defaultProps = new(ApplicationProps)
		defaultProps.init()
}

func GetSupportRunModes() []string {
		return supportRunModes
}

func (this *ApplicationProps) init() *ApplicationProps {
		this.initKeyValues()
		this.initProviders()
		return this
}

func (this *ApplicationProps) GetArgs() []string {
		return os.Args
}

func (this *ApplicationProps) GetProviders() []Contracts.Provider {
		return this.Providers
}

func (this *ApplicationProps) initKeyValues() {
		this.AppName = appName
		this.Version = appVersionName
		this.CtrChan = make(chan int, 2)
		this.ApkPath = reflect.TypeOf(this).Elem().PkgPath()
		this.ConfigFilesSuffix = []string{".yml", ".properties", ".ini"}
		this.BasePath = this.getCurrentDir()
		this.RunMode = this.getCurrentMode()
		this.ConfigDir = this.getCurrentConfigDir()
}

func (this *ApplicationProps) getCurrentDir() string {
		dir, _ := filepath.Abs(".")
		return filepath.Dir(dir)
}

func (this *ApplicationProps) getCurrentMode() string {
		mode := os.Getenv(RunModelEnv)
		if mode == "" {
				return RunModeDev
		}
		return mode
}

func (this *ApplicationProps) getCurrentConfigDir() string {
		dir := os.Getenv(ConfigDirEnv)
		if dir == "" {
				return this.BasePath + string(filepath.Separator) + "configs"
		}
		return dir
}

func (this *ApplicationProps) initProviders() {
		this.Providers = []Contracts.Provider{
				Components.AppBootstrapperOf(),         // bootstrapper
				Components.CommandLineArgsProviderOf(), // commandLine
				Components.EnvironmentProviderOf(),     // environment
				Components.ConfigureProviderOf(),       // configure
				Components.LoggerProviderOf(),          // logger
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
		case "appCtrlChan":
				fallthrough
		case "appctrlchan":
				fallthrough
		case "app_ctrl_chan":
				return this.CtrChan
		case "BasePath":
				fallthrough
		case "basepath":
				fallthrough
		case "base_path":
				return this.BasePath
		case "ConfigDir":
				fallthrough
		case "configdir":
				fallthrough
		case "config_dir":
				return this.ConfigDir
		case "RunMode":
				fallthrough
		case "runmode":
				fallthrough
		case "run_mode":
				return this.RunMode
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
				"ApkPath", "ConfigFilesSuffix", "appCtrlChan",
				"BasePath", "ConfigDir", "RunMode",
		}
}
