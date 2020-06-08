package Components

import (
		"encoding/json"
		"fmt"
		"reflect"
		"regexp"
		"strconv"
		"strings"
)

func ArrayOf(items ...interface{}) []interface{} {
		var arr []interface{}
		for _, it := range items {
				arr = append(arr, it)
		}
		return arr
}

func Array(item interface{}) []interface{} {
		var (
				values *reflect.Value
				arr    []interface{}
		)
		if reflect.TypeOf(item).Kind() == reflect.Array || reflect.TypeOf(item).Kind() == reflect.Slice {
				value := reflect.ValueOf(item)
				values = &value
		}
		if values == nil {
				el := reflect.ValueOf(item).Elem()
				if el.Kind() == reflect.Array || el.Kind() == reflect.Slice {
						values = &el
				}
		}
		if values != nil {
				count := values.Len()
				for i := 0; i < count; i++ {
						value := values.Index(i).Interface()
						arr = append(arr, value)
				}
				return arr
		}
		if item != nil {
				arr = append(arr, item)
		}
		return arr
}

func ArrayInclude(arr []interface{}, v interface{}) bool {
		var (
				ok bool
				it interface{}
				vc ComparisonEqual
		)
		vc, ok = v.(ComparisonEqual)
		for _, it = range arr {
				if it == v {
						return true
				}
				if ok && vc != nil {
						if vc.Equal(it) {
								return true
						}
				}
				if c, yes := it.(ComparisonEqual); yes {
						if c.Equal(v) {
								return true
						}
				}

		}
		return false
}

type ComparisonEqual interface {
		Equal(v interface{}) bool
}

func ArrayString(array []interface{}, all ...bool) []string {
		var arr []string
		if len(all) == 0 {
				all = append(all, false)
		}
		if len(array) == 0 {
				return arr
		}
		for _, v := range array {
				if str, ok := v.(string); ok {
						arr = append(arr, str)
						continue
				}
				if str, ok := v.(*string); ok {
						arr = append(arr, *str)
						continue
				}
				if all[0] {
						return arr
				}
				if str, ok := v.(fmt.Stringer); ok {
						arr = append(arr, str.String())
						continue
				}
				arr = append(arr, fmt.Sprintf("%v", v))
		}
		return arr
}

type IntegerArray interface {
		Invalid() bool
		ValueOf() []int
		fmt.Stringer
		Cap() int
		Count() int
}

type IntegerArrayImpl struct {
		value   interface{}
		cache   []int
		invalid int
}

func IntArray(args ...interface{}) IntegerArray {
		var arr = new(IntegerArrayImpl)
		arr.cache = []int{}
		arr.invalid = -1
		if len(args) > 0 {
				arr.Parse(args...)
		}
		return arr
}

func IntArrayOf(num ...int) IntegerArray {
		var arr = new(IntegerArrayImpl)
		arr.cache = []int{}
		arr.invalid = -1
		if len(num) > 0 {
				arr.invalid = 2
				for _, v := range num {
						arr.cache = append(arr.cache, v)
				}
		}
		return arr
}

func (this *IntegerArrayImpl) Parse(args ...interface{}) {
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

func (this *IntegerArrayImpl) ValueOf() []int {
		if !this.Invalid() {
				return this.cache
		}
		return []int{}
}

func (this *IntegerArrayImpl) Invalid() bool {
		if this.invalid != -1 {
				if this.invalid > 0 {
						return false
				}
				return true
		}

		return this.invalid <= 0
}

func (this *IntegerArrayImpl) parse() {
		if this.value == nil {
				this.invalid = 0
				return
		}
		switch this.value.(type) {
		case []int:
				this.invalid = 2
				this.cache = this.value.([]int)
		case *[]int:
				this.invalid = 2
				this.cache = *(this.value.(*[]int))
		case IntegerArray:
				bn := this.value.(IntegerArray)
				if !bn.Invalid() {
						this.invalid = 1
						this.cache = this.ValueOf()
				} else {
						this.invalid = 0
				}
		case string:
				this.invalid = 0
				if this.ParseStr(this.value.(string)) {
						this.invalid = 1
				}
		case *string:
				this.invalid = 0
				if this.ParseStr(*(this.value.(*string))) {
						this.invalid = 1
				}
		default:
				this.invalid = 0
		}
		return
}

func (this *IntegerArrayImpl) ParseStr(str string) bool {
		var (
				arr    []int
				arrStr []string
				reg2   = regexp.MustCompile(`^(([0-9]+)(,| )?)+$`)
				reg3   = regexp.MustCompile(`^\[(([0-9]+)(,| )?)+\]$`)
		)
		// dotArr 1,2,3,4
		// 1 23 343 55 6
		if reg2.MatchString(str) {
				if strings.Contains(str, ",") {
						arrStr = strings.SplitN(str, ",", -1)
				} else {
						arrStr = strings.SplitN(str, " ", -1)
				}
		}
		// [1,2,3,3] [1 2 2 3]
		if reg3.MatchString(str) {
				str = strings.Replace(str, "[", "", 1)
				str = strings.Replace(str, "]", "", 1)
				if strings.Contains(str, ",") {
						arrStr = strings.SplitN(str, ",", -1)
				} else {
						arrStr = strings.SplitN(str, " ", -1)
				}
		}

		if len(arrStr) > 0 {
				for _, it := range arrStr {
						it = strings.TrimSpace(it)
						if !IsNumber(it) {
								return false
						}
						n, err := strconv.Atoi(it)
						if err != nil {
								return false
						}
						arr = append(arr, n)
				}
				if len(arr) < 0 {
						return false
				}
				this.cache = arr
				return true
		}
		if err := json.Unmarshal([]byte(str), &arr); err == nil {
				this.cache = arr
				return true
		}
		return false
}

func (this *IntegerArrayImpl) Count() int {
		if !this.Invalid() {
				return len(this.ValueOf())
		}
		return 0
}

func (this *IntegerArrayImpl) Cap() int {
		if !this.Invalid() {
				return cap(this.ValueOf())
		}
		return 0
}

func (this *IntegerArrayImpl) String() string {
		return "<IntegerArray:" + reflect.TypeOf(this.value).Kind().String() + "> " + this.ToString()
}

func (this *IntegerArrayImpl) ToString() string {
		if this.value == nil {
				return "nil"
		}
		if !this.Invalid() {
				return fmt.Sprintf("%v", this.ValueOf())
		}
		return fmt.Sprintf("%v", this.value)
}
