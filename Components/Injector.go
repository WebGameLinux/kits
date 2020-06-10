package Components

import (
		"encoding/json"
		"errors"
		"github.com/fatih/structs"
		"github.com/mitchellh/mapstructure"
		"reflect"
)

type Injector struct {
		tags       []string
		MaxTry     int
		err        error
		DefaultTag string
}

var (
		DefaultInjectorTags = []string{
				"json", "inject", "mapstructure", "toml",
				"yml", "ini", "env", "service", "class", "var", "head", "key",
		}
)

// 注入器
func NewInjector(tags ...string) *Injector {
		var obj = new(Injector)
		if len(tags) > 0 {
				obj.tags = tags
		} else {
				obj.tags = DefaultInjectorTags
		}
		obj.MaxTry = 3
		obj.DefaultTag = "json"
		return obj
}

// 给对象注入属性
// source map[string]interface{}
// dist  <struct>
// tag dist 中的对应tag
func (this *Injector) Copy(source interface{}, dist interface{}, tags ...string) bool {
		if this.GetMapperType(source) == 0 && this.GetMapperType(dist) == 0 {
				return this.MapperCopy(source, dist)
		}
		if !this.checkDist(dist) {
				return false
		}
		if len(tags) == 0 {
				tags = this.tags
		}
		var count = 0
		// 解码到
		for _, tag := range tags {
				if this.MaxTry <= count {
						this.err = errors.New("times out failed,last tag at :" + tag)
						return false
				}
				this.err = this.Decode(source, dist, tag)
				if this.err == nil {
						return true
				}
				count++
		}
		if this.err == nil {
				this.err = errors.New("times out failed,last tag at :" + tags[count])
		}
		return false
}

func (this *Injector) GetMapperType(v interface{}) int {
		typ := reflect.TypeOf(v)
		if typ.Kind() != reflect.Map && typ.Kind() == reflect.Ptr {
				if typ.Elem().Kind() == reflect.Map {
						if _, ok := v.(*map[string]interface{}); ok {
								return 0
						}
						return 1
				}
				return -1
		}
		if _, ok := v.(map[string]interface{}); ok {
				return 0
		}
		return 1
}

func (this *Injector) MapperCopy(from, to interface{}) bool {
		var (
				source map[string]interface{}
				dist   map[string]interface{}
		)
		if v, ok := from.(map[string]interface{}); ok {
				source = v
		}
		if v, ok := to.(map[string]interface{}); ok {
				dist = v
		}
		if source == nil {
				if v, ok := from.(*map[string]interface{}); ok {
						source = *v
				}
		}
		if dist == nil {
				if v, ok := to.(*map[string]interface{}); ok {
						dist = *v
				}
		}
		if source != nil && dist != nil {
				for key, v := range source {
						dist[key] = v
				}
				return true
		}
		return false
}

// 解码,反序列化
func (this *Injector) Decode(input, output interface{}, tag ...string) error {
		if len(tag) == 0 {
				tag = append(tag, this.DefaultTag)
		}
		if b, ok := input.([]byte); ok {
				return json.Unmarshal(b, output)
		}
		config := &mapstructure.DecoderConfig{
				Metadata: nil,
				Result:   output,
				TagName:  tag[0],
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
				return err
		}
		return decoder.Decode(input)
}

func (this *Injector) Values(obj interface{}) map[string][]string {
		if !structs.IsStruct(obj) {
				return nil
		}
		var (
				count  = 0
				mapper = make(map[string][]string)
		)
		// 扫描tag
		for _, tag := range this.tags {
				if this.MaxTry <= count {
						break
				}
				structObj := structs.New(obj)
				structObj.TagName = tag
				arr := structObj.Names()
				if len(arr) == 0 {
						continue
				}
				count++
				mapper[tag] = arr
		}
		return mapper
}

func (this *Injector) checkDist(dist interface{}) bool {
		if dist == nil {
				this.err = errors.New("type error,dist type must be ptr, cause dist typeof nil")
				return false
		}
		typ := reflect.TypeOf(dist)
		if typ.Kind() != reflect.Ptr {
				this.err = errors.New("type error,dist type must be ptr, cause dist type:" + typ.Kind().String())
				return false
		}
		return true
}

func (this *Injector) Error() error {
		return this.err
}
