package main

import (
	"os"

	"github.com/wellcom-rocks/updates-sucks/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
