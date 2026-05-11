package main

import (
	"os"

	"github.com/maastrich/gh-next/cmd"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
