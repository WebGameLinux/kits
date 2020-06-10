package Clients

import (
		"fmt"
		"github.com/webGameLinux/kits/Libs/Errors"
		"net/url"
		"reflect"
		"strings"
)

type RpcMethod interface {
		Method() string // ${schema}://${user}@${password}/${host}/${path}/${action}/?options=xxx#anchor
		Encode() (string, error)
		Decode(...interface{}) (RpcMethod, error)
		Get(string) string
		fmt.Stringer
}

type KeysInterface interface {
		Keys() []string
}

const (
		MethodSchema    = "schema"
		MethodUser      = "user"
		MethodAuth      = "password"
		MethodHost      = "host"
		MethodAction    = "action"
		MethodPath      = "path"
		MethodQuery     = "?"
		MethodFragment  = "#"
		MethodExtras    = "_extras"
		MethodOpaque    = "opaque"
		MethodGenerator = "UrlGenerator"
		MethodUrl       = "Url"
		MethodProvider  = "MethodProvider"
)

var (
		SupportMethodKeys = []string{
				MethodSchema,
				MethodUser,
				MethodAuth,
				MethodHost,
				MethodAction,
				MethodPath,
				MethodQuery,
				MethodFragment,
				MethodOpaque,
				MethodExtras,
		}
)

type RpcMethodImpl struct {
		method  RpcMethod
		uri     string
		storage map[string]interface{}
}

type RpcMethodFactory func() RpcMethod
type UrlGenerator func(interface{}) string
type RpcMethodFilter func(interface{}) RpcMethod
type RpcMethodWrapper func(RpcMethod) RpcMethod
type UrlBalanceGenerator func(map[string]interface{}) string
type RpcMethodConstructor func(...interface{}) RpcMethod
type RpcMethodGetter func(map[string]interface{}) RpcMethod

func RpcMethodOf(args ...interface{}) RpcMethod {
		var method = new(RpcMethodImpl)
		if len(args) != 0 {
				method.init(args...)
		}
		return method
}

func (this *RpcMethodImpl) Method() string {
		method := this.getMethod()
		if method == this || method == nil {
				return this.getUrl()
		}
		return method.Method()
}

func (this *RpcMethodImpl) getUrl() string {
		if this.uri == "" {
				return this.uri
		}
		if generator, ok := this.storage[MethodGenerator]; ok {
				if gen, ok := generator.(UrlGenerator); ok {
						return gen(this.storage)
				}
				if gen, ok := generator.(UrlBalanceGenerator); ok {
						return gen(this.storage)
				}
		}
		if uri, ok := this.storage[MethodUrl]; ok {
				if str, ok := uri.(string); ok {
						this.uri = str
				}
		}
		if this.uri != "" {
				this.parse()
		}
		return this.uri
}

func (this *RpcMethodImpl) provider() RpcMethod {
		if instance, ok := this.storage[MethodProvider]; ok {
				if method, ok := instance.(RpcMethod); ok {
						return method
				}
				if method, ok := instance.(RpcMethodFactory); ok {
						m := method()
						if m != nil {
								this.method = m
						}
						return this.method
				}
				if constructor, ok := instance.(RpcMethodConstructor); ok {
						m := constructor(this.storage)
						if m != nil {
								this.method = m
						}
						return this.method
				}
				if getter, ok := instance.(RpcMethodGetter); ok {
						m := getter(this.storage)
						if m != nil {
								this.method = m
						}
						return this.method
				}
		}
		return this
}

func (this *RpcMethodImpl) Encode() (string, error) {
		method := this.getMethod()
		if method == nil || method == this {
				if this.uri != "" {
						return this.uri, nil
				}
				uri := this.toString()
				if uri != "" {
						return uri, nil
				}
				return "", Errors.TypeError("url nil encode")
		}
		return method.Encode()
}

func (this *RpcMethodImpl) Decode(args ...interface{}) (RpcMethod, error) {
		if len(args) == 0 {
				return this, nil
		}
		fn := args[0]
		// filter
		if filter, ok := fn.(RpcMethodFilter); ok {
				m := filter(this)
				if m == nil {
						return nil, Errors.TypeError("rpc filter failed")
				}
				return m, nil
		}
		// wrapper
		if wrapper, ok := fn.(RpcMethodWrapper); ok {
				m := wrapper(this)
				if m == nil {
						return nil, Errors.TypeError("rpc filter failed")
				}
				return m, nil
		}
		return RpcMethodOf(args...), nil
}

func (this *RpcMethodImpl) Get(key string) string {
		method := this.getMethod()
		if method == this {
				if len(this.storage) == 0 || this.uri != "" {
						this.parse()
				}
				switch key {
				case MethodSchema:
						fallthrough
				case MethodUser:
						fallthrough
				case MethodAuth:
						fallthrough
				case MethodHost:
						fallthrough
				case MethodAction:
						fallthrough
				case MethodPath:
						fallthrough
				case MethodQuery:
						fallthrough
				case MethodFragment:
						fallthrough
				case MethodExtras:
						fallthrough
				case MethodOpaque:
						return this.get(key)
				}
				return ""
		}
		return method.Get(key)
}

func (this *RpcMethodImpl) String() string {
		if uri, err := this.Encode(); err == nil {
				return uri
		}
		return ""
}

func (this *RpcMethodImpl) parse() {
		uri := this.uri
		if uri == "" {
				return
		}
		if Url, err := url.Parse(uri); err == nil {
				this.appendUrlParams(Url)
		}
}

func (this *RpcMethodImpl) toString() string {
		var uri = &url.URL{
				Path:     this.get(MethodPath),
				Scheme:   this.get(MethodSchema),
				Host:     this.get(MethodHost),
				Opaque:   this.get(MethodOpaque),
				RawQuery: this.get(MethodQuery),
				Fragment: this.get(MethodFragment),
		}
		if uri.RawQuery != "" {
				uri.ForceQuery = true
		}
		uri.User = url.UserPassword(this.get(MethodUser), this.get(MethodAuth))
		return uri.String()
}

func (this *RpcMethodImpl) get(key string) string {
		if v, ok := this.storage[key]; ok {
				if str, ok := v.(string); ok {
						return str
				}
				if str, ok := v.(fmt.Stringer); ok {
						return str.String()
				}
		}
		return ""
}

func (this *RpcMethodImpl) appendUrlParams(url *url.URL) {
		this.storage[MethodPath] = url.Path
		this.storage[MethodHost] = url.Host
		this.storage[MethodSchema] = url.Scheme
		this.storage[MethodFragment] = url.Fragment
		this.storage[MethodOpaque] = url.Opaque
		this.storage[MethodUser] = url.User.Username()
		// action
		if _, ok := this.storage[MethodAction]; !ok {
				this.storage[MethodAction] = this.getAction(url.Path)
		}
		if url.ForceQuery {
				this.storage[MethodQuery] = url.RawQuery
		}
		this.storage[MethodAuth], _ = url.User.Password()
}

func (this *RpcMethodImpl) getAction(path string) string {
		if strings.Contains(path, `/`) {
				paths := strings.SplitN(path, `/`, -1)
				return paths[len(paths)-1]
		}
		return path
}

func (this *RpcMethodImpl) Keys() []string {
		method := this.getMethod()
		if method == this {
				return SupportMethodKeys
		}
		if k, ok := method.(KeysInterface); ok {
				if _, ok := k.(*RpcMethodImpl); ok {
						return SupportMethodKeys
				}
				return k.Keys()
		}
		return SupportMethodKeys
}

func (this *RpcMethodImpl) getMethod() RpcMethod {
		if this.method != nil {
				return this.method
		}
		method := this.provider()
		if method != nil {
				return method
		}
		return this
}

func (this *RpcMethodImpl) init(args ...interface{}) {
		for _, arg := range args {
				if m, ok := arg.(RpcMethod); ok && this.method == nil {
						this.method = m
						continue
				}
				if m, ok := arg.(url.URL); ok && this.uri == "" {
						this.uri = m.String()
						continue
				}
				if m, ok := arg.(*url.URL); ok && this.uri == "" {
						this.uri = m.String()
						continue
				}
				if factory, ok := arg.(RpcMethodFactory); ok {
						this.storage[MethodProvider] = factory
						continue
				}
				if constructor, ok := arg.(RpcMethodConstructor); ok {
						this.storage[MethodProvider] = constructor
						continue
				}
				if getter, ok := arg.(RpcMethodGetter); ok {
						this.storage[MethodProvider] = getter
						continue
				}
				if urlGen, ok := arg.(UrlGenerator); ok {
						this.storage[MethodUrl] = urlGen
						continue
				}
				if method, ok := arg.(*RpcMethodImpl); ok {
						this.copy(method)
						continue
				}
				if mapper, ok := arg.(map[string]string); ok {
						this.mergeArgs(mapper)
						continue
				}
				if mapper, ok := arg.(map[string]interface{}); ok {
						this.mergeMapper(mapper)
						continue
				}
				values, ok := arg.(string)
				if ok && this.entry(values) {
						continue
				}
				if ok && values != "" && this.uri == "" && this.matchUrl(values) {
						this.uri = values
						continue
				}
				ty := reflect.TypeOf(arg)
				if ty.Kind() == reflect.Func {
						this.storage[ty.Kind().String()] = arg
				}
		}
}

func (this *RpcMethodImpl) mergeMapper(mapper map[string]interface{}) {
		for key, value := range mapper {
				v, ok := this.storage[key]
				if m, ok := v.(RpcMethod); ok && this.method == nil {
						this.method = m
				}
				if ok && reflect.TypeOf(value).Kind() == reflect.TypeOf(v).Kind() {
						this.storage[key] = value
				}
				if !ok {
						this.storage[key] = value
				}
		}
}

func (this *RpcMethodImpl) mergeArgs(mapper map[string]string) {
		for key, value := range mapper {
				v, ok := this.storage[key]
				if _, ok1 := v.(string); ok && ok1 && v != "" {
						this.storage[key] = value
				}
		}
}

func (this *RpcMethodImpl) entry(value string) bool {
		obj := EntryParse(value)
		if obj != nil && obj.Key() != "" {
				this.storage[obj.Key()] = obj.Value()
				return true
		}
		if this.uri == "" && value != "" && this.matchUrl(value) {
				this.uri = value
		}
		return false
}

func (this *RpcMethodImpl) copy(method *RpcMethodImpl) {
		if this.uri == "" {
				this.uri = method.uri
		}
		if len(this.storage) == 0 {
				this.storage = method.storage
		} else {
				this.mergeMapper(method.storage)
		}
		if this.method == nil && method.method != nil {
				this.method = method.method
		}
}

// 是否url
func (this *RpcMethodImpl) matchUrl(url string) bool {
		// ${schema}://${user}@${password}/${host}/${path}/${action}/?options=xxx#anchor
		if strings.Count(url, "://") >= 1 && strings.Count(url, "/") > 3 {
				return true
		}
		return false
}
