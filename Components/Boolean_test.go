package Components

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestBooleanOf(t *testing.T) {
		var True = BooleanOf(true)
		var False = BooleanOf(false)
		Convey("Boolean Test", t, func() {
				So(True.ValueOf(), ShouldEqual, true)
				So(False.ValueOf(), ShouldEqual, false)
				So(BooleanOf(1).ValueOf(), ShouldEqual, true)
				So(BooleanOf(0).ValueOf(), ShouldEqual, false)
				So(BooleanOf().ValueOf(), ShouldEqual, false)
				So(BooleanOf().Invalid(), ShouldEqual, true)
				So(BooleanOf("1").ValueOf(), ShouldEqual, true)
				So(BooleanOf("On").ValueOf(), ShouldEqual, true)
				So(BooleanOf("ON").ValueOf(), ShouldEqual, true)
				So(BooleanOf("NO").ValueOf(), ShouldEqual, false)
				So(BooleanOf("no").ValueOf(), ShouldEqual, false)
				So(BooleanOf("No").ValueOf(), ShouldEqual, false)
				So(BooleanOf("yes").ValueOf(), ShouldEqual, true)
				So(BooleanOf("Yes").ValueOf(), ShouldEqual, true)
				So(BooleanOf("YES").ValueOf(), ShouldEqual, true)
		})
}
