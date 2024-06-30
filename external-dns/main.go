package main

import (
	"fmt"
	"os"

	"git.houseofkummer.com/lior/home-dns/external-dns/cmd"
)

func main() {
	rootCmd := cmd.CreateRootCommand()
	if err := rootCmd.Cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
