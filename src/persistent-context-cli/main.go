package main

import (
	"fmt"
	"os"

	"github.com/JaimeStill/persistent-context/persistent-context-cli/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}