package WeChat

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestGetInstance(t *testing.T) {
		var ins = GetInstance()
		Convey("Test WeChat Instance",t, func() {
				So(ins,ShouldNotBeNil)
				So(ins.Get("default"),ShouldNotBeNil)
		})
}
