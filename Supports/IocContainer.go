package Supports

import (
	"github.com/webGameLinux/kits/Contracts"
	"sync"
)

const (
	BIND             = "bind"
	SINGLETON        = "singleton"
	REAL_CLAZZ       = "clazz"
	SINGLETON_OBJECT = "singleton_object"
)

type ContainerApp interface {
	Contracts.Container
	Resolver(string) *entry
}

type ContainerImpl struct {
	items []*entry
	mutex sync.Mutex
}

// 子项
type entry struct {
	key    string
	value  interface{}
	extras *Extras
}

// 扩展信息
type Extras struct {
	sync.Map
}

// 扩展信息
func Extrasof() *Extras {
	var extras = new(Extras)
	return extras
}

// 子项
func EntryOf(args ...interface{}) *entry {
	var (
		k    interface{}
		v    interface{}
		argc = len(args)
		en   = new(entry)
	)
	en.key = ""
	en.value = nil
	en.extras = Extrasof()
	if argc > 0 {
		k = args[0]
		if argc > 1 {
			v = args[1]
		}
		if key, ok := k.(string); ok {
			en.key = key
		}
		if v != nil {
			en.value = v
		}
		if argc >= 3 && nil != args[2] {
			if extras, ok := args[2].(*Extras); ok {
				en.extras = extras
			}
		}
	}
	return en
}

// entry
func Entry(v interface{}) (*entry, bool) {
	if v == nil {
		return nil, false
	}
	if it, ok := v.(entry); ok {
		return &it, true
	}
	if it, ok := v.(*entry); ok {
		return it, true
	}
	return nil, false
}

// 容器
func Containerof(items ...*entry) ContainerApp {
	var container = new(ContainerImpl)
	if len(items) > 0 {
		for _, item := range items {
			container.items = append(container.items, item)
		}
	}
	return container
}

func (this *Extras) Bool(key string) bool {
	if b, ok := this.Load(key); ok {
		if v, ok := b.(bool); ok {
			return v
		}
	}
	return false
}

func (this *Extras) String(key string) string {
	if b, ok := this.Load(key); ok {
		if v, ok := b.(string); ok {
			return v
		}
	}
	return ""
}

// key
func (this *entry) Key() string {
	return this.key
}

// 扩展信息
func (this *entry) Extras() *Extras {
	return this.extras
}

// 值
func (this *entry) Value() interface{} {
	return this.value
}

// 更正key
func (this *entry) SetKey(key string) *entry {
	this.key = key
	return this
}

// 更新 value
func (this *entry) SetValue(v interface{}) *entry {
	this.value = v
	return this
}

func (this *ContainerImpl) Get(id string) interface{} {
	for _, en := range this.items {
		if en.key == id {
			return en
		}
	}
	return nil
}

func (this *ContainerImpl) Alias(clazz string, alias string) {
	en := this.Resolver(clazz)
	if en == nil {
		return
	}
	extras := Extrasof()
	if class, ok := en.extras.Load(REAL_CLAZZ); ok {
		extras.Store(REAL_CLAZZ, class)
	} else {
		extras.Store(REAL_CLAZZ, en.value)
	}
	this.add(EntryOf(alias, clazz, extras))
}

func (this *ContainerImpl) Bind(id string, object interface{}) {
	if !this.Exists(id) {
		it := EntryOf(id, object)
		it.extras.Store(BIND, true)
		this.add(it)
	}
}

func (this *ContainerImpl) Singleton(id string, factory func(app Contracts.ApplicationContainer) interface{}) {
	if this.Exists(id) {
		return
	}
	it := EntryOf(id, factory)
	it.extras.Store(SINGLETON, true)
	this.add(it)
}

func (this *ContainerImpl) Destroy(ids ...string) {
	if len(ids) == 0 {
		this.items = []*entry{}
	}
}

func (this *ContainerImpl) Keys() []string {
	var keys []string
	for _, en := range this.items {
		keys = append(keys, en.Key())
	}
	return keys
}

func (this *ContainerImpl) Exists(id string) bool {
	for _, en := range this.items {
		if en.Key() == id {
			return true
		}
	}
	return false
}

func (this *ContainerImpl) Resolver(id string) *entry {
	en := this.Get(id)
	if en == nil {
		return nil
	}
	if obj, ok := en.(*entry); ok {
		return obj
	}
	return nil
}

func (this *ContainerImpl) add(it *entry) *ContainerImpl {
	this.items = append(this.items, it)
	return this
}
