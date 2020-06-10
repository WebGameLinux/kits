package Clients

import (
		"encoding/json"
		"fmt"
		"github.com/webGameLinux/kits/Libs/Errors"
		"reflect"
		"strings"
)

type RpcClientInterface interface {
		Call(method RpcMethod, params []RpcParam) RpcResult
}

type RpcClientImpl struct {
		client  RpcClientInterface
		storage map[string]interface{}
}

type EntryObject interface {
		Key() string
		Value() interface{}
}

// entry
type EntryObjectImpl struct {
		key   string
		value interface{}
}

type EntryPublicImpl struct {
		PubKey   string      `json:"key" entry:"key"`
		PubValue interface{} `json:"value" entry:"value"`
}

type RpcResult interface {
		Error() error
		Value() interface{}
		Bind(v interface{}) error
		Decode([]byte) (RpcResult, error)
		Encode(args ...interface{}) ([]byte, error)
}

// client 工厂函数
type ClientFactory func() RpcClientInterface
type ClientMethodFactory func(RpcMethod) RpcClientInterface
type ResultEncoder func(RpcResult) ([]byte, error)
type ResultFormatter func([]byte) []byte

type RpcResultImpl struct {
		Err  error       `json:"err" result:"err"`
		Data interface{} `json:"data" result:"data"`
}

const (
		RpcClientId     = "rpc-client"
		MockRpcClientId = "mock-rpc-client"
		//	ResultTag       = "result"
)

func RpcResultOf(args ...interface{}) RpcResult {
		var result = new(RpcResultImpl)
		if len(args) != 0 {
				result.init(args...)
		}
		return result
}

func RpcClientOf(args ...interface{}) RpcClientInterface {
		var client = new(RpcClientImpl)
		client.storage = make(map[string]interface{})
		if len(args) != 0 {
				client.init(args...)
		}
		return client
}

func (this *RpcResultImpl) Error() error {
		return this.Err
}

func (this *RpcResultImpl) Value() interface{} {
		return this.Value
}

func (this *RpcResultImpl) Bind(v interface{}) error {
		if v == nil {
				return Errors.NilPointError("bind object")
		}
		ty := reflect.TypeOf(v)
		if ty.Kind() != reflect.Ptr {
				return Errors.TypeError("bind object must struct ptr")
		}
		if encode, err := this.Encode(); err == nil {
				err := json.Unmarshal(encode, v)
				if err != nil {
						return Errors.UnmarshalError(err.Error())
				}
		}
		return nil
}

func (this *RpcResultImpl) Decode(data []byte) (RpcResult, error) {
		var result = RpcResultOf()
		err := json.Unmarshal(data, result)
		if err == nil {
				return result, nil
		}
		return result, err
}

func (this *RpcResultImpl) Encode(args ...interface{}) ([]byte, error) {
		var (
				err  error
				data []byte
				argc = len(args)
		)
		if argc == 0 {
				return json.Marshal(this)
		}
		if argc == 1 {
				if encoder, ok := args[0].(ResultEncoder); ok {
						return encoder(this)
				}
				if formatter, ok := args[0].(ResultFormatter); ok {
						if en, err := json.Marshal(this); err == nil {
								return formatter(en), nil
						}
				}
		}
		// 自定义 encode 和 格式化
		for _, v := range args {
				if encoder, ok := v.(ResultEncoder); ok && data == nil {
						data, err = encoder(this)
						if err != nil {
								data = nil
						}
				}
				if formatter, ok := v.(ResultFormatter); ok && data != nil {
						data = formatter(data)
				}
		}
		if data == nil || len(data) == 0 {
				if err != nil {
						return nil, Errors.UnmarshalError(err.Error())
				}
				return nil, Errors.UnmarshalError("unmarshal failed")
		}
		return data, nil
}

func (this *RpcResultImpl) init(args ...interface{}) {
		for _, arg := range args {
				err, ok := arg.(error)
				if ok && this.Err == nil {
						this.Err = err
						continue
				}
				res, ok := arg.(RpcResult)
				if ok && res != nil {
						if this.Data == nil {
								this.Data = res.Value()
						}
						if this.Err == nil {
								this.Err = res.Error()
						}
						continue
				}
				if arg != nil && this.Data == nil {
						this.Data = arg
						continue
				}
		}
}

// client
func (this *RpcClientImpl) init(args ...interface{}) {
		for _, v := range args {
				if client, ok := v.(RpcClientInterface); ok && this.client == nil {
						this.client = client
						continue
				}
				if entry, ok := v.(EntryObject); ok && entry != nil {
						this.storage[entry.Key()] = entry.Value()
				}
		}
}

func (this *RpcClientImpl) Call(method RpcMethod, params []RpcParam) RpcResult {
		client := this.getClient(method)
		if client != nil {
				return client.Call(method, params)
		}
		return RpcResultOf(Errors.NilClientError("client miss"))
}

func (this *RpcClientImpl) getClient(method RpcMethod) RpcClientInterface {
		// client
		if this.client != nil {
				return this.client
		}
		if len(this.storage) == 0 {
				return nil
		}
		value := this.storage[RpcClientId]
		// rpc client
		if client, ok := value.(RpcClientInterface); ok {
				this.client = client
				return client
		}
		// client factory
		if factory, ok := value.(ClientFactory); ok {
				client := factory()
				if client != nil {
						this.client = client
						return client
				}
		}
		// 指定服务的  client
		if factory, ok := value.(ClientMethodFactory); ok {
				client := factory(method)
				if client != nil {
						return client
				}
		}
		mock := this.storage[MockRpcClientId]
		// mock client
		if client, ok := mock.(RpcClientInterface); ok {
				return client
		}
		// mock factory client
		if factory, ok := mock.(ClientFactory); ok {
				return factory()
		}
		// 指定服务的 mock client
		if factory, ok := mock.(ClientMethodFactory); ok {
				client := factory(method)
				if client != nil {
						return client
				}
		}
		return nil
}

func EntryOf(args ...interface{}) EntryObject {
		var entry = new(EntryObjectImpl)
		if len(args) != 0 {
				entry.init(args...)
		}
		return entry
}

func (this *EntryObjectImpl) init(args ...interface{}) {
		if len(args) == 1 {
				var obj EntryObject
				if value, ok := args[0].(string); ok {
						obj = EntryParse(value)
				}
				if obj != nil && this.key == "" && this.value == nil {
						this.key = obj.Key()
						this.value = obj.Value()
						return
				}
		}
		for _, v := range args {
				if entry, ok := v.(*EntryPublicImpl); ok && this.key == "" && this.value == nil {
						this.key = entry.PubKey
						this.value = entry.PubValue
						continue
				}
				if entry, ok := v.(EntryObject); ok && this.key == "" && this.value == nil {
						this.key = entry.Key()
						this.value = entry.Value()
						continue
				}
				if str, ok := v.(string); ok && this.key == "" && str != "" {
						this.key = str
						continue
				}
				if this.value == nil && v != nil {
						this.value = v
				}
		}
}

func (this *EntryObjectImpl) Key() string {
		return this.key
}

func (this *EntryObjectImpl) Value() interface{} {
		return this.value
}

func (this *EntryObjectImpl) GetType() reflect.Kind {
		return reflect.TypeOf(this.value).Kind()
}

func (this *EntryObjectImpl) IsEmpty() bool {
		return this.Key() == "" && this.Value() == nil
}

func (this *EntryObjectImpl) Encode() ([]byte, error) {
		v, err := this.values()
		if err != nil {
				return []byte(""), err
		}
		data := fmt.Sprintf(`{"%s":"%s"}`, this.Key(), v)
		return []byte(data), nil
}

func (this *EntryObjectImpl) values() (string, error) {
		if this.Value() == nil {
				return "null", nil
		}
		return "", Errors.UnmarshalError("entry Data kind " + this.GetType().String())
}

func (this *EntryObjectImpl) String() string {
		v, _ := this.values()
		return fmt.Sprintf(`%s:%s`, this.Key(), v)
}

func (this *EntryObjectImpl) Decode(data []byte) EntryObject {
		if len(data) == 0 {
				return nil
		}
		value := new(EntryPublicImpl)
		if err := json.Unmarshal(data, value); err != nil {
				return EntryOf(value)
		}
		return nil
}

func (this *EntryPublicImpl) Key() string {
		return this.PubKey
}

func (this *EntryPublicImpl) Value() interface{} {
		return this.PubValue
}

func NewEntry(key string, value interface{}) EntryObject {
		var entry = EntryOf(key, value)
		return entry
}

func EntryParse(value string) EntryObject {
		var (
				key    string
				arr    []string
				values string
		)
		if strings.Contains(value, "{") && strings.Contains(value, "}") {
				v := make(map[string]interface{})
				if err := json.Unmarshal([]byte(value), &v); err == nil {
						k, ok := v["key"]
						val, ok2 := v["Data"]
						if key, ok3 := k.(string); ok3 && ok && ok2 {
								obj := new(EntryPublicImpl)
								obj.PubKey = key
								obj.PubValue = val
								return obj
						}
				}
		}
		if !strings.Contains(value, ":") {
				return nil
		}
		arr = strings.SplitN(value, ":", 1)
		if len(arr) < 2 {
				return nil
		}
		key, values = arr[0], arr[1]
		var entry = EntryOf(key, values)
		return entry
}
