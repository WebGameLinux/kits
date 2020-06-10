package main

import (
		. "github.com/webGameLinux/kits/Functions"
)

func main() {
		// 获取应用
		var app = AppContainer()
		// 引导加载
		Bootstrap(app)
		// 启动应用
		app.StarUp()
}
