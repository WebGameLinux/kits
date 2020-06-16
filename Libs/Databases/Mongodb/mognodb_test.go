package Mongodb

import (
		. "github.com/smartystreets/goconvey/convey"
		"gopkg.in/mgo.v2/bson"
		"testing"
		"time"
)

func TestNewMog(t *testing.T) {
		var m = NewMog()
		Convey("Mongodb Test", t, func() {
				c := m.Conn().DB("mongodb").C("user")
				So(c.Insert(bson.M{
						"name":       "mongodb",
						"lang":       "en",
						"age":        5,
						"created_at": time.Now(),
				}), ShouldBeNil)
				So(c, ShouldNotBeNil)
		})
}
