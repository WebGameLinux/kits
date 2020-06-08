package Components

import (
		"fmt"
		"reflect"
		"strconv"
		"strings"
)

type Number interface {
		FloatN() float64
		Float() float32
		Int() int
		IntN() int64
		fmt.Stringer
		NaN() bool
		Uint() uint
		UintN() uint64
		ValueOf() interface{}
}

type NumberImpl struct {
		nan        int
		value      interface{}
		kind       reflect.Kind
		cache      interface{}
		formatters []func(num interface{}) interface{}
}

func (this *NumberImpl) FloatN() float64 {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Float64 {
						v, ok := this.ValueOf().(float64)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseFloat(n.ToString()); ok {
						switch n.cache.(type) {
						case float64:
								return n.cache.(float64)
						case float32:
								return float64(n.cache.(float32))
						}
				}
		}
		return 0
}

func (this *NumberImpl) Float() float32 {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Float32 {
						v, ok := this.ValueOf().(float32)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseFloat(n.ToString(), 10, 32); ok {
						v := n.ValueOf()
						switch v.(type) {
						case float64:
								return float32(v.(float64))
						case float32:
								return v.(float32)
						}
				}
		}
		return 0
}

func (this *NumberImpl) Int() int {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Int {
						v, ok := this.ValueOf().(int)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseInt(n.ToString()); ok {
						v := n.ValueOf()
						switch v.(type) {
						case int:
								return v.(int)
						case int32:
								return int(v.(int32))
						case int64:
								return int(v.(int64))
						}
				}
		}
		return 0
}

func (this *NumberImpl) IntN() int64 {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Int64 {
						v, ok := this.ValueOf().(int64)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseInt(n.ToString(), 10, 64); ok {
						v := n.ValueOf()
						switch v.(type) {
						case int:
								return int64(v.(int))
						case int32:
								return int64(int(v.(int32)))
						case int64:
								return v.(int64)
						}
				}
		}
		return 0
}

// uint64
func (this *NumberImpl) UintN() uint64 {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Uint64 {
						v, ok := this.ValueOf().(uint64)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseUint(n.ToString(), 10, 64); ok {
						v := n.ValueOf()
						switch v.(type) {
						case uint:
								return uint64(v.(uint))
						case uint16:
								return uint64(v.(uint16))
						case uint32:
								return uint64(int(v.(uint32)))
						case uint64:
								return v.(uint64)
						}
				}
		}
		return 0
}

// uint
func (this *NumberImpl) Uint() uint {
		if this.NaN() {
				return 0
		}
		if this.nan == 2 {
				if this.kind == reflect.Uint {
						v, ok := this.ValueOf().(uint)
						if ok {
								return v
						}
						return 0
				}
		}
		num := NumberOf(this.ToString())
		if n, ok := num.(*NumberImpl); ok {
				if n.ParseUint(n.ToString()); ok {
						v := n.ValueOf()
						switch v.(type) {
						case uint:
								return v.(uint)
						case uint16:
								return uint(v.(uint16))
						case uint32:
								return uint(int(v.(uint32)))
						case uint64:
								return uint(v.(uint64))
						}
				}
		}
		return 0
}

func (this *NumberImpl) ValueOf() interface{} {
		if this.cache == nil {
				return this.value
		}
		return this.cache
}

// 是否非数字
func (this *NumberImpl) NaN() bool {
		if this.nan != -1 {
				if this.nan > 0 {
						return false
				}
				return true
		}

		switch this.value.(type) {
		case uint:
				this.nan = 2
				this.kind = reflect.Uint
		case uint8:
				this.nan = 2
				this.kind = reflect.Uint8
		case uint16:
				this.nan = 2
				this.kind = reflect.Uint16
		case uint32:
				this.nan = 2
				this.kind = reflect.Uint32
		case uint64:
				this.nan = 2
				this.kind = reflect.Uint64
		case int:
				this.nan = 2
				this.kind = reflect.Int
		case int8:
				this.nan = 2
				this.kind = reflect.Int8
		case int16:
				this.nan = 2
				this.kind = reflect.Int16
		case int32:
				this.nan = 2
				this.kind = reflect.Int32
		case int64:
				this.nan = 2
				this.kind = reflect.Int64
		case float32:
				this.nan = 2
				this.kind = reflect.Float32
		case float64:
				this.nan = 2
				this.kind = reflect.Float64
		case string:
				this.kind = reflect.String
				if IsNumber(this.value.(string)) {
						this.nan = 1
						this.ParseStr()
				} else {
						this.nan = 0
						if len(this.formatters) == 0 {
								this.ParseStr()
						}
				}
		case Number:
				if v, ok := this.value.(Number); ok {
						if v.NaN() {
								this.nan = 0
						} else {
								this.nan = 1
						}
						this.cache = v.ValueOf()
						this.kind = reflect.TypeOf(v).Kind()
				}
		default:
				this.nan = 0
				if this.value == nil {
						this.kind = reflect.Invalid
				} else {
						this.kind = reflect.TypeOf(this.value).Kind()
				}
		}
		return this.nan <= 0
}

// 非法类型
func (this *NumberImpl) Invalid() bool {
		return this.NaN()
}

// arg0 - num|formatter0
// arg1 - formatter0
// argN - formatterN
func (this *NumberImpl) Parse(args ...interface{}) {
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
		for _, fn := range args {
				if formatter, ok := fn.(func(interface{}) interface{}); ok {
						this.formatters = append(this.formatters, formatter)
				}
		}
		for _, formatter := range this.formatters {
				if formatter != nil && this.value != nil && this.value != "" {
						this.value = formatter(this.value)
				}
		}
		this.NaN()
}

// 解析字符串
func (this *NumberImpl) ParseStr() {
		if str, ok := this.value.(string); ok {
				// 1000,000,000
				if strings.Contains(str, ","); ok {
						str = strings.Replace(str, ",", "", -1)
				}
				// 小数
				hasDot := strings.Contains(str, ".")
				// 负数
				hasOpt := strings.Contains(str, "-")
				// 1 000 000
				str = strings.TrimSpace(str)
				// float
				if hasDot && this.ParseFloat(str) {
						return
				}
				// unit
				if !hasOpt && !hasDot && this.ParseUint(str) {
						return
				}
				// int
				if !hasDot && this.ParseInt(str) {
						return
				}
				// default
				if i, err := strconv.Atoi(str); err == nil {
						this.cache = i
						this.nan = 1
						this.kind = reflect.Int
						return
				}
		}
}

// 解析 字符串 float
func (this *NumberImpl) ParseFloat(str string, bitSize ...int) bool {
		if len(bitSize) != 0 && (bitSize[0] == 64 || bitSize[0] == 32) {
				if i, err := strconv.ParseFloat(str, bitSize[0]); err == nil {
						this.cache = i
						this.nan = 1
						this.kind = reflect.TypeOf(this.cache).Kind()
						return true
				}
				return false
		}
		if i, err := strconv.ParseFloat(str, 64); err == nil {
				this.cache = i
				this.kind = reflect.Float64
				this.nan = 1
				return true
		}
		if i, err := strconv.ParseFloat(str, 32); err == nil {
				this.cache = i
				this.kind = reflect.Float32
				this.nan = 1
				return true
		}
		return false
}

// 解析 字符串 uint
func (this *NumberImpl) ParseUint(str string, bitSize ...int) bool {
		if len(bitSize) >= 2 && (bitSize[1] == 64 || bitSize[1] == 32 || bitSize[1] == 16 || bitSize[1] == 8) {
				if i, err := strconv.ParseUint(str, bitSize[0], bitSize[1]); err == nil {
						this.cache = i
						this.nan = 1
						this.kind = reflect.TypeOf(this.cache).Kind()
						return true
				}
				return false
		}
		if i, err := strconv.ParseUint(str, 10, 64); err == nil {
				this.cache = i
				this.kind = reflect.Uint64
				this.nan = 1
				return true
		}
		if i, err := strconv.ParseUint(str, 10, 32); err == nil {
				this.cache = i
				this.kind = reflect.Uint32
				this.nan = 1
				return true
		}
		if i, err := strconv.ParseUint(str, 10, 16); err == nil {
				this.cache = i
				this.kind = reflect.Uint16
				this.nan = 1
				return true
		}
		if i, err := strconv.ParseUint(str, 10, 8); err == nil {
				this.cache = i
				this.kind = reflect.Uint8
				this.nan = 1
				return true
		}
		return false
}

// 解析 字符串 int
func (this *NumberImpl) ParseInt(str string, bitSize ...int) bool {
		if len(bitSize) >= 2 && (bitSize[1] == 64 || bitSize[1] == 32) {
				if i, err := strconv.ParseInt(str, bitSize[0], bitSize[1]); err == nil {
						this.cache = i
						this.nan = 1
						this.kind = reflect.TypeOf(this.cache).Kind()
						return true
				}
				return false
		}
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
				this.cache = i
				this.nan = 1
				this.kind = reflect.Int64
				return true
		}
		if i, err := strconv.ParseInt(str, 10, 32); err == nil {
				this.cache = i
				this.nan = 1
				this.kind = reflect.Int64
				return true
		}
		return false
}

// 字符串接口
func (this *NumberImpl) String() string {
		if !this.NaN() {
				return "<Number:" + this.kind.String() + "> " + this.ToString()
		}
		return "<Number:" + this.kind.String() + "> NaN"
}

// 获取字符串数据
func (this *NumberImpl) ToString() string {
		v := this.ValueOf()
		if str, ok := v.(fmt.Stringer); ok {
				return str.String()
		}
		if v == nil {
				return "nil"
		}
		return fmt.Sprintf("%v", v)
}

// 数字解析
// arg0 - num|formatter0
// arg1 - formatter0
// argN - formatterN
func NumberOf(any ...interface{}) Number {
		var num = new(NumberImpl)
		num.nan = -1
		num.cache = nil
		num.formatters = []func(interface{}) interface{}{}
		if len(any) > 0 {
				num.Parse(any...)
		}
		return num
}
