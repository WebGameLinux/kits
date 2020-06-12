package test

import (
		"fmt"
		"go/importer"
		"testing"
)

func TestScript(t *testing.T)  {
		include := importer.Default()
		fmt.Println(include)
}
