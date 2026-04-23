package main

import (
	"fmt"
	"os"

	"github.com/lukewilliamboswell/libasciidoc/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCmd("ascii2doc", "docx")
	rootCmd.AddCommand(cli.NewVersionCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
