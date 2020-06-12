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

## 应用配置配置支持

.env  和 运行环境 ${mode}.env 

支持 指定 单个配置 配置文件加载 和 配置文件夹加载

支持 yml , toml , ini , properties 配置格式 

支持 配置中 键值 环境变量表达式 ge : ``${version}``

支持 配置中  键值 环境变量带默认值表达式 eg:  ``$(version|1.0.0)``  

支持 应用 debug 调试配置 ``app_debug=true`` (env文件中指定或者环境变量中指定)  

