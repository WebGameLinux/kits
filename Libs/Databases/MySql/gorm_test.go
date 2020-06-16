package MySql

import (
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestNewMysqlConnector(t *testing.T) {

		var connector = NewMysqlConnector(Connection{
				User: "root",
				Password: "root@123",
				Database: "admin",
		})

		Convey("test convey", t, func() {
				db := connector.Conn()
				So(db, ShouldNotBeNil)
		})
}
