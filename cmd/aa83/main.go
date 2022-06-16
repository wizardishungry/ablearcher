package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

func main() {
	bi, _ := debug.ReadBuildInfo()
	fmt.Println(bi.Main)
	e, _ := os.Executable()
	fmt.Println(e)
	// flag.NewFlagSet()
}
