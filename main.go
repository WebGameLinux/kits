package main

import (
	"fmt"
	. "github.com/webGameLinux/kits/Supports"
)

func main() {
	var app = App()
	app.Alias(AppContainer, "container")
	fmt.Printf("%T\n", app.Get("container"))
	fmt.Printf("%T\n", app.Get(AppContainer))
	app.StarUp()
}
