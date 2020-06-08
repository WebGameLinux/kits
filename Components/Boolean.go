package Components

import (
		"fmt"
		"reflect"
)

type Boolean interface {
		Invalid() bool
		ValueOf() bool
		fmt.Stringer
}

type BooleanImpl struct {
		value   interface{}
		cache   bool
		invalid int
}

func BooleanOf(args ...interface{}) Boolean {
		var b = new(BooleanImpl)
		b.cache = false
		b.invalid = -1
		if len(args) != 0 {
				b.Parse(args...)
		}
		return b
}

func (this *BooleanImpl) Invalid() bool {
		if this.invalid != -1 {
				if this.invalid > 0 {
						return false
				}
				return true
		}
		this.parse()
		return this.invalid <= 0
}

func (this *BooleanImpl) ValueOf() bool {
		if !this.Invalid() {
				return this.cache
		}
		if b, ok := this.value.(bool); ok {
				return b
		}
		return false
}

func (this *BooleanImpl) String() string {
		return "<Boolean:" + reflect.TypeOf(this.value).String() + "> " + this.ToString()
}

func (this *BooleanImpl) ToString() string {
		return fmt.Sprintf("%v", this.ValueOf())
}

// 解析和自定义格式器
// arg0  value
// arg1  formatter0
// argN  formatterN
func (this *BooleanImpl) Parse(args ...interface{}) {
		if len(args) <= 0 {
				return
		}
		if args[0] == nil {
				return
		}
		_, ok := args[0].(func(interface{}) string)
		if !ok {
				if this.value == nil {
						this.value = args[0]
				}
		}
		var formatters []func(interface{}) interface{}
		for _, fn := range args {
				if formatter, ok := fn.(func(interface{}) interface{}); ok {
						formatters = append(formatters, formatter)
				}
		}
		for _, formatter := range formatters {
				if formatter != nil && this.value != nil && this.value != "" {
						this.value = formatter(this.value)
				}
		}
		this.parse()
}

func (this *BooleanImpl) parse() {
		switch this.value.(type) {
		case bool:
				this.cache = this.value.(bool)
				this.invalid = 2
		case int:
				b := this.value.(int)
				if b == 1 {
						this.invalid = 1
						this.cache = true
				}
				if b == 0 {
						this.invalid = 1
						this.cache = false
				}
		case uint:
				b := this.value.(uint)
				if b == 1 {
						this.invalid = 1
						this.cache = true
				}
				if b == 0 {
						this.invalid = 1
						this.cache = false
				}
		case string:
				str := this.value.(string)
				this.invalid = -1
				if this.IsTrue(str) {
						this.cache = true
						this.invalid = 1
				}
				if this.IsFalse(str) {
						this.cache = false
						this.invalid = 1
				}
		default:
				this.invalid = -1
		}
}

func (this *BooleanImpl) IsTrue(str string) bool {
		return ArrayInclude(trueArr, str)
}

func (this *BooleanImpl) IsFalse(str string) bool {
		return ArrayInclude(falseArr, str)
}

var (
		trueArr  = Array([]string{"true", "True", "TRUE", "1", "on", "On", "ON", "yes", "Yes", "YES", "up", "Up", "UP", "ok", "Ok", "OK"})
		falseArr = Array([]string{"false", "False", "FALSE", "0", "off", "Off","OFF", "no", "No","NO", "down", "Down","DOWN"})
)
