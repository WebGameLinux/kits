package Supports

import (
		"fmt"
		"github.com/webGameLinux/kits/Contracts"
		"io"
		"os"
		"strings"
		"sync"
)

type Properties struct {
		paths       []string
		appFile     string
		reader      io.Reader
		init        bool
		commandStop bool
		options     map[string]map[string]string
		cache       map[string]string
}

var (
		propertiesInstanceLock sync.Once
		propertiesInstance     *Properties
		optionsMapper          = map[string][]string{
				"appFile": {"-f", "--file", "-c", "@tip:主配置文件 @eg:-f /path/app.properties "},
				"paths":   {"-p", "--paths", "@tip:配置目录 @eg:--paths=/paths,/path2"},
				"reader":  {"-r", "--reader", "@tip:读取器 @eg: --reader=/paths/a.ini"},
				"mode":    {"-m", "--mode", "@tip:运行环境 @eg: --mode=test"},
				"help":    {"-h", "--help", "@@"},
		}
)

const (
		HelpStop = Contracts.HelpStop
)

func AppBasePropertiesOf() *Properties {
		if propertiesInstance == nil {
				propertiesInstanceLock.Do(newProperties)
		}
		return propertiesInstance
}

func GetOptions() map[string][]string {
		return optionsMapper
}

func newProperties() {
		propertiesInstance = new(Properties)
		propertiesInstance.paths = []string{}
		propertiesInstance.options = map[string]map[string]string{}
}

func (this *Properties) GetReader() io.Reader {
		return this.reader
}

func (this *Properties) Keys() []string {
		var (
				defaults = []string{"reader", "appFile", "paths", "cStop"}
				mapper   = this.GetOptions()
		)
		if len(mapper) != 0 {
				var (
						keys  []string
						cache = make(map[string]bool)
				)
				// 用户自定义参数
				for key, _ := range mapper {
						if _, ok := cache[key]; ok {
								continue
						}
						keys = append(keys, key)
						cache[key] = true
				}
				// 默认
				for _, key := range defaults {
						if _, ok := cache[key]; ok {
								continue
						}
						keys = append(keys, key)
						cache[key] = true
				}
				return keys
		}
		return defaults
}

func (this *Properties) get(key string) string {
		for k, v := range this.GetOptions() {
				if k == key || strings.EqualFold(key, k) {
						return v
				}
		}
		return ""
}

func (this *Properties) Values() []interface{} {
		return []interface{}{
				this.paths,
				this.appFile,
				this.reader,
				this.commandStop,
		}
}

func (this *Properties) Mapper() map[string]interface{} {
		return map[string]interface{}{
				"paths":   this.paths,
				"appFile": this.appFile,
				"reader":  this.reader,
				"cStop":   this.commandStop,
		}
}

func (this *Properties) Inited() bool {
		return this.init
}

func (this *Properties) Get(key string) interface{} {
		switch key {
		case "appFile":
				fallthrough
		case "appfile":
				fallthrough
		case "App.Properties.File":
				return this.appFile
		case "App.Properties.files":
				return strings.SplitN(this.appFile, ",", -1)
		case "paths":
				fallthrough
		case "Paths":
				fallthrough
		case "App.Properties.Paths":
				return this.paths
		case "reader":
				fallthrough
		case "Reader":
				fallthrough
		case "App.Properties.Reader":
				return this.reader
		case "inited":
				fallthrough
		case "isInited":
				return this.init
		case "cStop":
		case HelpStop:
				return this.commandStop
		}
		if v := this.get(key); v != "" {
				return v
		}
		return nil
}

func (this *Properties) SetReader(reader io.Reader) *Properties {
		if this.reader == nil && !this.init {
				this.reader = reader
		}
		return this
}

func (this *Properties) SetFile(file string) *Properties {
		if this.appFile == "" && !this.init {
				this.appFile = file
		}
		return this
}

func (this *Properties) SetPaths(paths []string) *Properties {
		if len(this.paths) == 0 && !this.init {
				this.paths = paths
		}
		return this
}

func (this *Properties) AppendPath(paths []string) *Properties {
		if !this.init {
				for _, pa := range paths {
						ok := true
						for _, p := range this.paths {
								if p == pa {
										ok = false
										break
								}
						}
						if ok {
								this.paths = append(this.paths, pa)
						}
				}

		}
		return this
}

func (this *Properties) Init() {
		this.initEnv()
		this.initArgs()
		this.init = true
		this.help()
}

func (this *Properties) help() {
		if h, ok := this.GetOptions()["help"]; ok && h != "" {
				this.menu()
				this.stop()
		}
}

func (this *Properties) stop() {
		this.commandStop = true
		this.cache[HelpStop] = "true"
		this.options["cStop"] = map[string]string{"ok": "true"}
}

func (this *Properties) menu() {
		fmt.Println("  commander help options :")
		fmt.Println("  		commander [options] ")
		for key, arr := range GetOptions() {
				v := arr[len(arr)-1]
				if key == "help" {
						v = strings.Replace(v, "@@", "help show menu", 1)
				} else {
						v = strings.Replace(v, "@tip:", " ", 1)
				}
				arr[len(arr)-1] = strings.Replace(v, "@eg:", "eg: ", 1)
				fmt.Printf("		 %s :  %s \n", key, strings.Join(arr, " "))
		}
}

func (this *Properties) initEnv() {
		this.loaderEnv()
		for key, val := range this.GetOptions() {
				this.set(key, val)
		}
}

func (this *Properties) initArgs() {
		this.parse()

		for key, val := range this.GetOptions() {
				this.set(key, val)
		}
}

func (this *Properties) set(key string, value string) {
		switch key {
		case "appFile":
				fallthrough
		case "appfile":
				fallthrough
		case Contracts.AppPropertiesFile:
				fallthrough
		case Contracts.AppPropertiesFiles:
				this.appFile = value
		case "paths":
				fallthrough
		case "Paths":
				fallthrough
		case Contracts.AppPropertiesPaths:
				this.paths = strings.SplitN(value, ",", -1)
		case "reader":
				fallthrough
		case "Reader":
				fallthrough
		case Contracts.AppPropertiesReader:
				if this.reader != nil {
						return
				}
				this.reader = this.newReader(value)
		case "inited":
				fallthrough
		case "isInited":
				if value == "1" || value == "true" {
						this.init = true
				} else {
						this.init = false
				}
		case "cStop":
		case HelpStop:
				if value == "1" || value == "true" {
						this.commandStop = true
				} else {
						this.commandStop = false
				}
		}
}

func (this *Properties) newReader(fs string) io.Reader {
		if strings.Index(fs, "$") == 0 {
				tmp := this.env(fs)
				if tmp == "" {
						return nil
				}
				fs = tmp
		}
		if state, err := os.Stat(fs); err == nil {
				if state.IsDir() {
						return nil
				}
		}
		reader, err := os.Open(fs)
		if err != nil {
				return nil
		}
		return reader
}

func (this *Properties) env(str string) string {
		return ParseEnvStr(str)
}

func (this *Properties) Foreach(each func(k string, v interface{}) bool) {
		for _, key := range this.Keys() {
				exportKey := this.With(key)
				value := this.Get(exportKey)
				if !each(exportKey, value) {
						break
				}
		}
}

func (this *Properties) With(key string) string {
		switch key {
		case "appFile":
				return "App.Properties.files"
		case "reader":
				return "App.Properties.Reader"
		case "paths":
				return "App.Properties.Paths"
		case "cStop":
				return HelpStop
		}
		return key
}

func (this *Properties) GetOptions() map[string]string {
		if this.cache == nil || len(this.cache) != len(this.options) {
				var mapper = make(map[string]string)
				for k, m := range this.options {
						for _, v := range m {
								mapper[k] = v
								break
						}
				}
				this.cache = mapper
		}
		return this.cache
}

func (this *Properties) Configure(loaderInterface Contracts.PropertyLoaderInterface) func(k string, v interface{}) bool {
		return func(k string, v interface{}) bool {
				if loaderInterface == nil {
						return false
				}
				loaderInterface.PropertyLoader(func(s *sync.Map) {
						if k == Contracts.ArgRunMode && IsSupportMode(v) {
								s.Store(Contracts.RunModeEnv, strings.ToLower(v.(string)))
								return
						}
						s.Store(k, v)
				})
				return true
		}
}

func (this *Properties) parse() {
		if this.init {
				return
		}
		var (
				val    string
				args   = os.Args
				values []string
				option []map[string]string
		)
		if len(args) < 2 {
				return
		}
		for _, arg := range args[1:] {
				tmp := []rune(arg)
				if len(arg) < 2 {
						values = this.appendOrPop(arg, values, &option)
						continue
				}
				if tmp[0] != '-' {
						values = this.appendOrPop(arg, values, &option)
						continue
				}
				val = ""
				if strings.Contains(arg, "=") {
						arr := strings.SplitN(arg, "=", 2)
						arg = arr[0]
						if len(arr) >= 2 {
								val = arr[1]
						}
				}
				opt, _ := this.FindOption(arg)
				if opt == "" {
						if val != "" {
								k := strings.Replace(arg, "-", "", 2)
								this.options[k] = map[string]string{arg: val}
								this.updateCache(k, val)
						}
						continue
				}
				if val != "" {
						this.options[opt] = map[string]string{arg: val}
						this.updateCache(opt, val)
				} else {
						option = append(option, map[string]string{opt: arg})
				}
		}
		// true value fill
		if len(option) > 0 {
				for _, it := range option {
						for k, v := range it {
								this.options[k] = map[string]string{v: "true"}
								this.updateCache(k, "true")
						}
				}
		}
}

func (this *Properties) updateCache(key string, val string) {
		if this.cache == nil {
				this.cache = make(map[string]string)
		}
		this.cache[key] = val
}

func (this *Properties) FindOption(k string) (string, []string) {
		k = strings.TrimSpace(k)
		for key, options := range GetOptions() {
				for _, opt := range options[0 : len(options)-1] {
						if k == opt {
								return key, options
						}
				}
		}
		return "", []string{}
}

func (this *Properties) loaderEnv() {
		for _, key := range this.Keys() {
				if key == "cStop" {
						continue
				}
				k := key
				if key == "appFile" {
						k = "app_file"
				}
				env := strings.ToUpper(k)
				v := os.Getenv(env)
				this.options[key] = map[string]string{env: v}
		}
}

// 出存
func (this *Properties) appendOrPop(arg string, arr []string, options *[]map[string]string) []string {
		arr = append(arr, arg)
		if len(*options) <= 0 {
				return arr
		}
		var (
				i     int
				count = len(arr)
				m     map[string]string
		)
		for i, m = range *options {
				if i >= count {
						break
				}
				for key, t := range m {
						this.options[key] = map[string]string{t: arr[i]}
						this.updateCache(key, arr[i])
				}
		}
		if i+1 == count {
				*options = (*options)[0:0]
				return arr[0:0]
		}
		*options = (*options)[i:]
		return arr[i:]
}

func ParseEnvStr(key string) string {
		var (
				count  int
				mapper = make(map[string]string)
		)
		for strings.Contains(key, "${") && strings.Contains(key, "}") {
				arr := strings.SplitN(key, "${", -1)
				for _, ky := range arr {
						if strings.Contains(ky, "}") {
								vars := strings.SplitN(ky, "}", 1)
								k := "${" + vars[0] + "}"
								val := os.Getenv(vars[0])
								if _, ok := mapper[k]; ok {
										if count > 1 {
												return key
										}
										continue
								}
								mapper[k] = val
								if val != "" {
										key = strings.Replace(key, k, val, -1)
								}
						}
				}
				count++
		}
		return key
}

func IsSupportMode(v interface{}) bool {
		var m string
		if str, ok := v.(string); ok {
				m = str
		}
		if str, ok := v.(*string); ok {
				m = *str
		}
		if str, ok := v.(fmt.Stringer); ok {
				m = str.String()
		}
		if m == "" {
				return false
		}
		for _, mode := range supportRunModes {
				if m == mode || strings.EqualFold(m, mode) {
						return true
				}
		}
		return false
}
