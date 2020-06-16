package Cache

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
		"time"
)

func TestRedis(t *testing.T) {
		var (
			  key = "redis"
			  value = "123"
				cache = Redis()
		)
		Convey("Redis Test", t, func() {
				So(cache.Conn(), ShouldNotBeNil)
				So(cache.Client(), ShouldNotBeNil)
				ctx := cache.Context()
				time.AfterFunc(100*time.Microsecond, func() {
						ctx.Deadline()
				})
				res := cache.Client().Set(ctx, key, value, 2*time.Second)
				So(IsSuccess(res), ShouldEqual, true)
				re :=cache.Client().Get(ctx,key)
				So(GetResult(re),ShouldEqual,value)
		})
}
