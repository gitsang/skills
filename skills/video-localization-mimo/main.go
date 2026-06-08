package main

import (
	"fmt"
	"os"

	"github.com/gitsang/skills/video-localization-mimo/cmd/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		fmt.Println("Use: go run cmd/server/main.go")
		os.Exit(1)
	} else {
		cli.Execute()
	}
}
