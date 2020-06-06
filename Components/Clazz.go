package Components

import (
	"fmt"
	"github.com/webGameLinux/kits/Contracts"
	"reflect"
)

// class 实现
type ClazzImpl struct {
	Name            string
	ConstructorFunc func() interface{}
	FactoryFunc     func(Contracts.ApplicationContainer) interface{}
}

// 自动生成 clazz
func ClazzOf(obj ...interface{}) Contracts.ClazzInterface {
	var clazz = new(ClazzImpl)
	if len(obj) > 0 {
		ClassOf(obj[0], clazz)
	}
	return clazz
}

func (this *ClazzImpl) String() string {
	return this.Name
}

func (this *ClazzImpl) Clazz() string {
	return this.Name
}

func (this *ClazzImpl) Constructor() func() interface{} {
	return this.ConstructorFunc
}

func (this *ClazzImpl) Factory() func(Contracts.ApplicationContainer) interface{} {
	return this.FactoryFunc
}

// 自动获取
func ClassOf(obj interface{}, clazz *ClazzImpl) bool {
	if obj == nil || clazz == nil {
		return false
	}
	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Struct {
		if t.Elem().Kind() != reflect.Struct {
			return false
		}
		t = t.Elem()
	}
	if class, ok := obj.(*ClazzImpl); ok {
		clazz.Name = class.Name
		clazz.ConstructorFunc = class.ConstructorFunc
		clazz.FactoryFunc = class.FactoryFunc
		return true
	}
	if t.Implements(reflect.TypeOf(new(Contracts.FactoryInterface))) {
		if factory, ok := obj.(Contracts.FactoryInterface); ok {
			clazz.FactoryFunc = factory.Factory
		}
	}
	if t.Implements(reflect.TypeOf(new(Contracts.ConstructorInterface))) {
		if factory, ok := obj.(Contracts.ConstructorInterface); ok {
			clazz.ConstructorFunc = factory.Constructor
		}
	}
	if t.Implements(reflect.TypeOf(new(fmt.Stringer))) {
		if factory, ok := obj.(fmt.Stringer); ok {
			clazz.Name = factory.String()
		}
	}
	if clazz.Name == "" {
		return false
	}
	if clazz.ConstructorFunc == nil && clazz.FactoryFunc == nil {
		return false
	}
	return true
}
