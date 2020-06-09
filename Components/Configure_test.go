package Components

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestConfigureOf(t *testing.T) {
		var cnf = ConfigureOf()
		cnf.Load("../config")
		Convey("Configure Test", t, func() {
				So(cnf.Get("app.app"), ShouldEqual, "kits")
				So(len(cnf.IntArray("app.arr")), ShouldEqual, len([]int{1, 2, 3, 67, 0, 0}))
		})
}
