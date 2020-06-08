package Components

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestArray(t *testing.T) {
		var arr1 = []interface{}{1, 2, 3, 3, 4, "false", NumberOf(123)}
		Convey("Array Test", t, func() {
				var arr = Array(&arr1)
				So(len(arr), ShouldEqual, len(arr1))
				for i, v := range arr {
						So(v, ShouldEqual, arr1[i])
				}
				Convey("One item", func() {
						var one = arr[len(arr)-1]
						var arr2 = Array(one)
						So(len(arr2), ShouldEqual, 1)
						So(arr2[0], ShouldEqual, one)
				})
		})
}

func TestIntArrayOf(t *testing.T) {
		var bInt = []int{1, 2, 3, 4,}
		Convey("IntegerArray Test", t, func() {
				var arr = IntArray(bInt)
				So(arr.Count(), ShouldEqual, len(bInt))
				var strArr = IntArray("1,2,3,4,5")
				So(strArr.Count(), ShouldEqual, 5)
				So(strArr.Invalid(), ShouldEqual, false)
				var strArr2 = IntArray("[1,2,3,4,5]")
				So(strArr2.Count(), ShouldEqual, 5)
				So(strArr2.Invalid(), ShouldEqual, false)
				var strArr3 = IntArray("[1 2 3 4 5]")
				So(strArr3.Count(), ShouldEqual, 5)
				So(strArr3.Invalid(), ShouldEqual, false)
				var strArr4 = IntArray("1 2 3 4 5")
				So(strArr4.Count(), ShouldEqual, 5)
				So(strArr4.Invalid(), ShouldEqual, false)
		})
}
