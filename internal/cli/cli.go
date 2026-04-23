package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/lukewilliamboswell/libasciidoc/asciidoc"
	"github.com/lukewilliamboswell/libasciidoc/configuration"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRootCmd returns a root command configured with the given default backend.
func NewRootCmd(name, defaultBackend string) *cobra.Command {

	var noHeaderFooter bool
	var outputName string
	var logLevel string
	var css []string
	var backend string
	var attributes []string
	var themePath string

	rootCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [flags] FILE", name),
		Short: fmt.Sprintf("%s converts AsciiDoc files to %s", name, defaultBackend),
		Args:  cobra.ArbitraryArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := log.ParseLevel(logLevel)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "unable to parse log level '%v'", logLevel)
				return err
			}
			log.SetFormatter(&log.TextFormatter{
				EnvironmentOverrideColors: true,
				DisableLevelTruncation:    true,
				DisableTimestamp:          true,
			})
			log.SetLevel(lvl)
			log.SetOutput(cmd.ErrOrStderr())
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			attrs := parseAttributes(attributes)
			for _, sourcePath := range args {
				out, close := getOut(cmd, sourcePath, outputName, backend)
				if out != nil {
					defer close() //nolint:errcheck
					config := configuration.NewConfiguration(
						configuration.WithFilename(sourcePath),
						configuration.WithAttributes(attrs),
						configuration.WithCSS(css),
						configuration.WithBackEnd(backend),
						configuration.WithHeaderFooter(!noHeaderFooter),
						configuration.WithThemePath(themePath))
					_, err := asciidoc.ConvertFile(out, config)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	rootCmd.SilenceUsage = true
	flags := rootCmd.Flags()
	flags.BoolVarP(&noHeaderFooter, "no-header-footer", "s", false, "do not render header/footer")
	flags.StringVarP(&outputName, "output", "o", "", "output file (default: based on path of input file); use - to output to STDOUT")
	flags.StringVar(&logLevel, "log-level", "warn", "log level to set [debug|info|warn|error|fatal|panic]")
	flags.StringArrayVar(&css, "css", []string{}, "the paths to the CSS files to link to the document")
	flags.StringArrayVarP(&attributes, "attribute", "a", []string{}, "a document attribute to set in the form of name, name!, or name=value pair")
	flags.StringVarP(&backend, "backend", "b", defaultBackend, "backend to format the file")
	flags.StringVar(&themePath, "theme", "", "path to Asciidoctor PDF theme YAML file for DOCX styling")
	return rootCmd
}

// NewVersionCmd returns the version command.
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

type closeFunc func() error

func defaultCloseFunc() closeFunc {
	return func() error { return nil }
}

func newCloseFileFunc(c io.Closer) closeFunc {
	return func() error {
		return c.Close()
	}
}

func getOut(cmd *cobra.Command, sourcePath, outputName, backend string) (io.Writer, closeFunc) {
	if outputName == "-" {
		return cmd.OutOrStdout(), defaultCloseFunc()
	} else if outputName != "" {
		outfile, err := os.Create(outputName)
		if err != nil {
			log.Warnf("Cannot create output file - %v, skipping", outputName)
			return nil, defaultCloseFunc()
		}
		return outfile, newCloseFileFunc(outfile)
	} else if sourcePath != "" {
		path, err := filepath.Abs(sourcePath)
		if err != nil {
			log.Warnf("Cannot resolve absolute path for %v: %v, skipping", sourcePath, err)
			return nil, defaultCloseFunc()
		}
		ext := ".html"
		if backend == "docx" {
			ext = ".docx"
		}
		outname := strings.TrimSuffix(path, filepath.Ext(path)) + ext
		outfile, err := os.Create(outname)
		if err != nil {
			log.Warnf("Cannot create output file - %v, skipping", outname)
			return nil, defaultCloseFunc()
		}
		return outfile, newCloseFileFunc(outfile)
	}
	return cmd.OutOrStdout(), defaultCloseFunc()
}

func parseAttributes(attributes []string) map[string]interface{} {
	result := make(map[string]interface{}, len(attributes))
	for _, attr := range attributes {
		data := strings.SplitN(attr, "=", 2)
		if len(data) > 1 {
			result[data[0]] = data[1]
		} else {
			result[data[0]] = ""
		}
	}
	return result
}
