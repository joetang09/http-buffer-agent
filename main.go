package main

import (
	"fmt"
	"os"

	"github.com/joetang09/http-buffer-agent/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
