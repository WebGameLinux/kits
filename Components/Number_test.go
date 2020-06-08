package Components

import (
		. "github.com/smartystreets/goconvey/convey"
		"math"
		"testing"
)

func TestNumberOf(t *testing.T) {
		Convey("Number Interface Test", t, func() {
				data := getNaNTestData()
				for v, e := range data {
						num := NumberOf(v)
						So(num.NaN(), ShouldEqual, e)
				}
		})
}

func getNaNTestData() map[interface{}]bool {
		return map[interface{}]bool{
				"123":         false,
				0:             false,
				"1.00":        false,
				"123x":        true,
				"abc":         true,
				"0o00oo00":    true,
				1.00:          false,
				NumberOf(123): false,
				math.MaxInt64: false,
				false:         true,
				nil:           true,
				"":            true,
		}
}
