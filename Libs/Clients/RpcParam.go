package Clients

import "encoding/json"

type RpcParam interface {
		Key() string        // key
		ParamType() string  // header, body , query , path,options
		Value() interface{} // Data
		Encode() ([]byte, error)
		Set(string, interface{})
		Null() bool
		Decode(arg ...interface{}) (RpcParam, error)
}

type RpcParamImpl struct {
		PType  string      `json:"type" param:"type"`
		PKey   string      `json:"key" param:"key"`
		PValue interface{} `json:"value" param:"value"`
}

const (
		ParamKey        = "key"
		ParamValue      = "value"
		ParamType       = "type"
		ParamTag        = "param"
		ParamTypeHeader = "header"
		ParamTypeQuery  = "query"
		ParamTypeBody   = "body"
		ParamTypePath   = "path"
		ParamTypeValue  = "value"
		ParamTypeVar    = "var"
)

var (
		ParamTypeSupports = []string{
				ParamTypeHeader,
				ParamTypeQuery,
				ParamTypeBody,
				ParamTypePath,
				ParamTypeValue,
				ParamTypeVar,
		}
)

type RpcParamArray []RpcParam

type RpcWrapper func(RpcParam) RpcParam

func RpcParamOf(args ...interface{}) RpcParam {
		var param = new(RpcParamImpl)
		if len(args) != 0 {
				param.init(args...)
		}
		return param
}

func (this *RpcParamImpl) Key() string {
		return this.PKey
}

func (this *RpcParamImpl) ParamType() string {
		return this.PType
}

func (this *RpcParamImpl) Value() interface{} {
		return this.PValue
}

func (this *RpcParamImpl) Encode() ([]byte, error) {
		return json.Marshal(this)
}

func (this *RpcParamImpl) Decode(args ...interface{}) (RpcParam, error) {
		var (
				argc = len(args)
		)
		if argc == 0 {
				return RpcParamOf(), nil
		}
		if argc == 1 {
				if wrapper, ok := args[0].(RpcWrapper); ok {
						return wrapper(this), nil
				}
		}
		return RpcParamOf(args...), nil
}

func (this *RpcParamImpl) Null() bool {
		return this.Key() == "" || this.Value() == nil
}

func (this *RpcParamImpl) init(args ...interface{}) {
		for _, v := range args {
				if p, ok := v.(RpcParam); ok {
						this.copy(p)
				}
				if str, ok := v.(string); ok {
						this.entry(str)
				}
				if mapper, ok := v.(map[string]interface{}); ok {
						this.mapperInit(mapper)
				}
				if mapper, ok := v.(*map[string]interface{}); ok {
						this.mapperInit(*mapper)
				}
		}
}

func (this *RpcParamImpl) copy(param RpcParam) {
		this.PValue = param.Value()
		this.PType = param.ParamType()
		this.PKey = param.Key()
}

func (this *RpcParamImpl) entry(val string) {
		obj := EntryParse(val)
		if obj.Key() != "" {
				this.mapperInit(map[string]interface{}{obj.Key(): obj.Value()})
				return
		}
		if this.PType == "" {
				this.PType = val
				return
		}
		if this.PKey == "" {
				this.PKey = val
				return
		}
		if this.PValue == "" {
				this.PValue = val
				return
		}
}

func (this *RpcParamImpl) mapperInit(mapper map[string]interface{}) {
		for key, v := range mapper {
				switch key {
				case ParamKey:
						if this.PKey != "" {
								return
						}
						this.Set(key, v)
				case ParamValue:
						if this.PValue != nil {
								return
						}
						this.Set(key, v)
				case ParamType:
						if this.PType != "" {
								return
						}
						this.Set(key, v)
				}
		}
}

func (this *RpcParamImpl) Set(key string, v interface{}) {
		switch key {
		case ParamKey:
				if str, ok := v.(string); ok {
						this.PKey = str
				}
		case ParamValue:
				if v != nil {
						this.PValue = v
				}
		case ParamType:
				if str, ok := v.(string); ok {
						this.PType = str
				}
		}
}

func (this RpcParamArray) Mapper() map[string][]RpcParam {
		var (
				i        int
				ok       bool
				key      string
				group    string
				indexKey string
				arr      []RpcParam
				index    = make(map[string]int)
				mapper   = make(map[string][]RpcParam)
		)
		for _, param := range this {
				key = param.Key()
				group = param.ParamType()
				indexKey = group + "." + key
				arr, ok = mapper[group]
				if !ok {
						mapper[group] = []RpcParam{param}
						continue
				}
				if i, ok = index[indexKey]; ok {
						if param.Value() == nil && arr[i].Value() != nil {
								continue
						}
						arr[i] = param
						continue
				}
				arr = append(arr, param)
				index[indexKey] = len(arr) - 1
		}
		return mapper
}

func (this RpcParamArray) Get(key string, ty ...string) interface{} {
		if len(ty) == 0 {
				ty = append(ty, ParamTypeVar)
		}
		var (
				val interface{}
		)
		for _, it := range this {
				if it.Key() == key {
						val = it.Value()
						if ty[0] == it.ParamType() {
								continue
						}
				}
		}
		return val
}

func (this RpcParamArray) Group(key string) map[string]interface{} {
		var mapper = make(map[string]interface{})
		for _, it := range this {
				if it.ParamType() != key {
						continue
				}
				mapper[it.Key()] = it.Value()
		}
		return mapper
}

func (this RpcParamArray) GroupKeys() []string {
		var (
				keys   []string
				mapper = make(map[string]int)
		)
		for _, it := range this {
				k := it.ParamType()
				v, ok := mapper[k]
				if !ok {
						mapper[k] = v + 1
				} else {
						mapper[k] = 1
				}
				if v == 1 {
						keys = append(keys, it.ParamType())
				}
		}
		return keys
}

func (this RpcParamArray) Set(ty, key string, value interface{}) RpcParamArray {
		var obj = RpcParamOf(map[string]interface{}{ParamType: ty, ParamKey: key, ParamValue: value})
		return this.Add(obj)
}

func (this RpcParamArray) String() string {
		if m, err := json.Marshal(this.Mapper()); err == nil {
				return string(m)
		}
		return ""
}

func (this RpcParamArray) Add(param RpcParam) RpcParamArray {
		if this.Update(param) {
				return this
		}
		return append(this, param)
}

func (this RpcParamArray) Update(param RpcParam) bool {
		for i, it := range this {
				if it.Key() == param.Key() && it.ParamType() == param.ParamType() {
						it.Set(ParamValue, param.Value())
						this[i] = it
						return true
				}
		}
		return false
}

func GetParamTypes() []string {
		return ParamTypeSupports
}
