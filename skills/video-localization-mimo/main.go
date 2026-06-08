package main

import (
	"fmt"
	"os"

	"github.com/gitsang/skills/video-localization-mimo/cmd/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		fmt.Println("Web 服务器模式尚未实现")
		os.Exit(1)
	} else {
		cli.Execute()
	}
}