package main

import (
	"fmt"
	"os"

	"github.com/sneakybugs/corewarden/api/cmd"
)

func main() {
	rootCmd := cmd.CreateRootCommand()
	if err := rootCmd.Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
