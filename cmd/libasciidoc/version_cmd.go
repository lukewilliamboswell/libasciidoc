package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// NewVersionCmd returns the version command
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version and build info",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			info, ok := debug.ReadBuildInfo()
			if !ok {
				fmt.Fprintln(out, "version: unknown (no build info)")
				return
			}
			printed := false
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				fmt.Fprintf(out, "version: %s\n", info.Main.Version)
				printed = true
			}
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					fmt.Fprintf(out, "commit:  %s\n", s.Value)
					printed = true
				case "vcs.time":
					fmt.Fprintf(out, "built:   %s\n", s.Value)
					printed = true
				case "vcs.modified":
					if s.Value == "true" {
						fmt.Fprintln(out, "dirty:   true")
						printed = true
					}
				}
			}
			if !printed {
				fmt.Fprintln(out, "version: development build")
			}
		},
	}
}
