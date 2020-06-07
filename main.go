package main

import (
		. "github.com/webGameLinux/kits/Supports"
)

func main() {
	var app = App()
	app.Alias(AppContainer, "container")
	app.StarUp()
}
