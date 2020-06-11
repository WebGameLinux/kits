package main

import (
		"fmt"
		. "github.com/webGameLinux/kits/Functions"
)

func main() {
		// 获取应用
		var app = AppContainer()
		// 引导加载
		Bootstrap(app)
		app.Bind("health", func() {
				fmt.Println(CnfKv("app.database.driver"))
		})
		// 启动应用
		app.StarUp()
}
