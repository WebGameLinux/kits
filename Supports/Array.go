package Supports

import (
		"fmt"
		"github.com/webGameLinux/kits/Contracts"
		"sync"
)

type RegisterUniqueArray interface {
		Contracts.UniqueArray
		Contracts.ArrayAggregation
		Start(index int) RegisterUniqueArray
		Value(interface{}) Contracts.RegisterInterface
}

type BooterUniqueArray interface {
		Contracts.UniqueArray
		Contracts.ArrayAggregation
		Start(index int) BooterUniqueArray
		Value(interface{}) Contracts.BootInterface
}

type RegisterUniqueArrayImpl struct {
		items []Contracts.RegisterInterface
		mutex sync.Mutex
}

type BooterUniqueArrayImpl struct {
		items []Contracts.BootInterface
		mutex sync.Mutex
}

func (this *RegisterUniqueArrayImpl) Exists(value interface{}) bool {
		if value == nil {
				return false
		}
		v := this.format(value)
		if v == nil {
				return false
		}
		for _, it := range this.items {
				if it == v {
						return true
				}
				if fmt.Sprintf("%s", it) == fmt.Sprintf("%s", v) {
						return true
				}
		}
		return false
}

func (this *RegisterUniqueArrayImpl) format(value interface{}) Contracts.RegisterInterface {
		v, ok := value.(Contracts.RegisterInterface)
		if !ok {
				return nil
		}
		return v
}

func (this *RegisterUniqueArrayImpl) Add(item interface{}) bool {
		v := this.format(item)
		if v == nil || this.Exists(item) {
				return false
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.items = append(this.items, v)
		return true
}

func (this *RegisterUniqueArrayImpl) Count() int {
		return len(this.items)
}

func (this *RegisterUniqueArrayImpl) Cap() int {
		return cap(this.items)
}

func (this *RegisterUniqueArrayImpl) OffsetGet(index interface{}) (interface{}, bool) {
		key := this.key(index)
		if key == -1 || this.Count() <= key {
				return nil, false
		}
		if v := this.items[key]; v == nil {
				return v, true
		}
		return nil, false
}

func (this *RegisterUniqueArrayImpl) key(v interface{}) int {
		if key, ok := v.(int); ok && key >= 0 {
				return key
		}
		return -1
}

func (this *RegisterUniqueArrayImpl) OffsetSet(key interface{}, value interface{}) {
		k := this.key(key)
		v := this.Value(value)
		if k < 0 || v == nil {
				return
		}
		if k >= this.Cap() {
				this.Add(v)
				return
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.items[k] = v
}

func (this *RegisterUniqueArrayImpl) OffsetExists(value interface{}) bool {
		key := this.key(value)
		if key < 0 || this.Count() < key {
				return false
		}
		return nil != this.items[key]
}

func (this *RegisterUniqueArrayImpl) OffsetUnset(k interface{}) {
		key := this.key(k)
		count := this.Count()
		if key < 0 || count <= key || count == 0 {
				return
		}
		var arr []Contracts.RegisterInterface
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if key == 0 {
				this.items = append(arr, this.items[1:]...)
		}
		if key > 0 && key < count {
				arr = append(arr, this.items[0:key]...)
				this.items = append(arr, this.items[key+1:]...)
		}
		if key+1 == count {
				this.items = this.items[0 : count-1]
		}
}

func (this *RegisterUniqueArrayImpl) Foreach(each func(key, value interface{}) bool) {
		var count = len(this.items)
		for i := 0; i < count; {
				if !each(i, this.items[i]) {
						break
				}
				cur := len(this.items)
				if count == cur {
						i++
				}
				if cur == 0 {
						break
				}
		}
}

func (this *RegisterUniqueArrayImpl) Empty() bool {
		if this.Count() == 0 {
				return true
		}
		arr := this.Filter(func(key, value interface{}) bool {
				return value != nil
		})
		return arr.Count() == 0
}

func (this *RegisterUniqueArrayImpl) Filter(filter func(key, value interface{}) bool) Contracts.ArrayAccess {
		var array = RegisterUniqueArrayOf()
		for k, v := range this.items {
				if filter(k, v) {
						array.Add(v)
				}
		}
		return array
}

func (this *RegisterUniqueArrayImpl) Stream(stream func(key, value interface{}) bool) Contracts.ArrayAggregation {
		var array = RegisterUniqueArrayOf()
		for k, v := range this.items {
				if stream(k, v) {
						array.Add(v)
				}
		}
		return array
}

func (this *RegisterUniqueArrayImpl) Value(v interface{}) Contracts.RegisterInterface {
		if boot, ok := v.(Contracts.RegisterInterface); ok {
				return boot
		}
		return nil
}

func (this *RegisterUniqueArrayImpl) Start(index int) RegisterUniqueArray {
		var (
				count = this.Count()
		)
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if count <= index || count < 0 {
				return RegisterUniqueArrayOf()
		}
		return RegisterUniqueArrayOf(this.items[index:]...)
}

func (this *RegisterUniqueArrayImpl) RPop() interface{} {
		var num = this.Count()
		if num < 0 {
				return nil
		}
		var (
				index = len(this.items) - 1
				it    = this.items[index]
		)
		this.items = this.items[0:index]
		return it
}

func (this *RegisterUniqueArrayImpl) LPop() interface{} {
		var num = this.Count()
		if num < 0 {
				return nil
		}
		var (
				index = 0
				start = 1
				it    = this.items[index]
		)
		if start < num {
				this.items = this.items[start:]
		} else {
				this.items = this.items[0:0]
		}
		return it
}

func (this *BooterUniqueArrayImpl) Exists(value interface{}) bool {
		if value == nil {
				return false
		}
		v := this.format(value)
		if v == nil {
				return false
		}
		for _, it := range this.items {
				if it == v {
						return true
				}
				if fmt.Sprintf("%s", it) == fmt.Sprintf("%s", v) {
						return true
				}
		}
		return false
}

func (this *BooterUniqueArrayImpl) Add(item interface{}) bool {
		v := this.format(item)
		if v == nil || this.Exists(item) {
				return false
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.items = append(this.items, v)
		return true
}

func (this *BooterUniqueArrayImpl) Count() int {
		return len(this.items)
}

func (this *BooterUniqueArrayImpl) Cap() int {
		return cap(this.items)
}

func (this *BooterUniqueArrayImpl) OffsetGet(index interface{}) (interface{}, bool) {
		key := this.key(index)
		if key == -1 || this.Count() <= key {
				return nil, false
		}
		if v := this.items[key]; v == nil {
				return v, true
		}
		return nil, false
}

func (this *BooterUniqueArrayImpl) OffsetSet(key interface{}, value interface{}) {
		k := this.key(key)
		v := this.Value(value)
		if k < 0 || v == nil {
				return
		}
		if k >= this.Cap() {
				this.Add(v)
				return
		}
		this.mutex.Lock()
		defer this.mutex.Unlock()
		this.items[k] = v
}

func (this *BooterUniqueArrayImpl) OffsetExists(value interface{}) bool {
		key := this.key(value)
		if key < 0 || this.Count() < key {
				return false
		}
		return nil != this.items[key]
}

func (this *BooterUniqueArrayImpl) OffsetUnset(k interface{}) {
		key := this.key(k)
		count := this.Count()
		if key < 0 || count <= key || count == 0 {
				return
		}
		var arr []Contracts.BootInterface
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if key == 0 {
				this.items = append(arr, this.items[1:]...)
		}
		if key > 0 && key < count {
				arr = append(arr, this.items[0:key]...)
				this.items = append(arr, this.items[key+1:]...)
		}
		if key+1 == count {
				this.items = this.items[0 : count-1]
		}
}

func (this *BooterUniqueArrayImpl) Foreach(each func(key, value interface{}) bool) {
		var count = len(this.items)
		for i := 0; i < count; {
				if !each(i, this.items[i]) {
						break
				}
				cur := len(this.items)
				if count == cur {
						i++
				}
				if cur == 0 {
						break
				}
		}
}

func (this *BooterUniqueArrayImpl) RPop() interface{} {
		var num = this.Count()
		if num < 0 {
				return nil
		}
		var (
				index = len(this.items) - 1
				it    = this.items[index]
		)
		this.items = this.items[0:index]
		return it
}

func (this *BooterUniqueArrayImpl) LPop() interface{} {
		var num = this.Count()
		if num < 0 {
				return nil
		}
		var (
				index = 0
				start = 1
				it    = this.items[index]
		)
		if start < num {
				this.items = this.items[start:]
		} else {
				this.items = this.items[0:0]
		}
		return it
}

func (this *BooterUniqueArrayImpl) Filter(filter func(key, value interface{}) bool) Contracts.ArrayAccess {
		var array = BooterUniqueArrayOf()
		for k, v := range this.items {
				if filter(k, v) {
						array.Add(v)
				}
		}
		return array
}

func (this *BooterUniqueArrayImpl) Empty() bool {
		if this.Count() == 0 {
				return true
		}
		arr := this.Filter(func(key, value interface{}) bool {
				return value != nil
		})
		return arr.Count() == 0
}

func (this *BooterUniqueArrayImpl) Stream(stream func(key, value interface{}) bool) Contracts.ArrayAggregation {
		var array = BooterUniqueArrayOf()
		for k, v := range this.items {
				if stream(k, v) {
						array.Add(v)
				}
		}
		return array
}

func (this *BooterUniqueArrayImpl) Start(index int) BooterUniqueArray {
		var (
				count = this.Count()
		)
		this.mutex.Lock()
		defer this.mutex.Unlock()
		if count <= index || count < 0 {
				return BooterUniqueArrayOf()
		}
		return BooterUniqueArrayOf(this.items[index:]...)
}

func (this *BooterUniqueArrayImpl) format(value interface{}) Contracts.BootInterface {
		v, ok := value.(Contracts.BootInterface)
		if !ok {
				return nil
		}
		return v
}

func (this *BooterUniqueArrayImpl) key(v interface{}) int {
		if key, ok := v.(int); ok && key >= 0 {
				return key
		}
		return -1
}

func (this *BooterUniqueArrayImpl) Value(v interface{}) Contracts.BootInterface {
		if boot, ok := v.(Contracts.BootInterface); ok {
				return boot
		}
		return nil
}

func RegisterUniqueArrayOf(items ...Contracts.RegisterInterface) RegisterUniqueArray {
		var array = RegisterUniqueArrayImpl{}
		if len(items) == 0 {
				return &array
		}
		for _, it := range items {
				array.Add(it)
		}
		return &array
}

func BooterUniqueArrayOf(items ...Contracts.BootInterface) BooterUniqueArray {
		var array = BooterUniqueArrayImpl{}
		if len(items) == 0 {
				return &array
		}
		for _, it := range items {
				array.Add(it)
		}
		return &array
}
