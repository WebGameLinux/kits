# kits for golang

简介 :

![info](./images/WechatIMG4.jpg)

项目 意在构建一个类 laravel的 golang 开发组件包

## 使用方式 

```bash
go get github.com/webGameLinux/kits
```

main.go 编写

```go
package main

import (
		. "github.com/webGameLinux/kits/Functions"
)

func main() {
		// 获取应用
		var app = AppContainer()
		// 引导加载,用户可以自定义引导
		Bootstrap(app)
		// 启动应用
		app.StarUp()
}

```
