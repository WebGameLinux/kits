package Etcd

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestNewConnector(t *testing.T) {
		var conn = NewConnector()
		Convey("Test Etcd Connector", t, func() {
				So(conn.Conn(), ShouldNotBeNil)
		})
}
