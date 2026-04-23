package testsupport

import (
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

func init() {
	lvl := parseLogLevel()
	log.SetLevel(lvl)
	log.Warnf("Running test with logs in '%s' level", lvl.String())
	log.SetFormatter(&log.TextFormatter{
		DisableQuote: true, // see https://github.com/sirupsen/logrus/issues/608#issuecomment-745137306
	})

	// also, configuration for spew (when dumping structures to compare results)
	spew.Config.DisableCapacities = true
	spew.Config.DisablePointerAddresses = true
	spew.Config.DisablePointerMethods = true
}

func parseLogLevel() log.Level {
	logLevel := "error"
	// Scan os.Args manually so that unknown flags (e.g. ginkgo's) are silently ignored.
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--loglevel" || arg == "-l":
			if i+1 < len(args) {
				logLevel = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--loglevel="):
			logLevel = arg[len("--loglevel="):]
		case strings.HasPrefix(arg, "-l="):
			logLevel = arg[len("-l="):]
		}
	}
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	return lvl
}
