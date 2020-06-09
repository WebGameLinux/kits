package Libs

import (
		"fmt"
		. "github.com/smartystreets/goconvey/convey"
		"testing"
)

func TestNewViperLoader(t *testing.T) {
		var cnf = NewViperLoader()
		cnf.Mapper.AddConfigPath("../Config")
		cnf.Mapper.SetConfigName("app.properties")
		cnf.Mapper.SetConfigType("properties")
		_=cnf.Mapper.ReadInConfig()
		// fmt.Println(err)
		Convey("Viper Config Test",t, func() {
				So(cnf.Mapper.Get("app.name"),ShouldEqual,"kits")
				cnf.Foreach(func(k, v interface{}) bool {
						fmt.Println(k,v)
						return true
				})
				So(cnf.Get("app.grpc.port"),ShouldEqual,cnf.Get("app.http.port"))
		})
}

func TestNewViperGet(t *testing.T) {
		var cnf = NewViperLoader()
		cnf.Mapper.AddConfigPath("../Config")
	  cnf.Mapper.SetConfigName("app.ini")
	  cnf.Mapper.SetConfigType("ini")
		_=cnf.Mapper.ReadInConfig()
		Convey("Viper ini Config Test",t, func() {
				So(cnf.Get("default.app"),ShouldEqual,"kits")

		})
}

func TestNewViperEnv(t *testing.T) {
		var cnf = NewViperLoader()
		cnf.Mapper.AddConfigPath("../Config")
		cnf.Mapper.SetConfigName("app.env")
		cnf.Mapper.SetConfigType("env")
		_=cnf.Mapper.ReadInConfig()
		Convey("Viper Env Config Test",t, func() {
				So(cnf.Get("app.name"),ShouldEqual,"kits")
		})
}
