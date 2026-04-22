module github.com/bytesparadise/libasciidoc

go 1.24.0

require (
	github.com/alecthomas/chroma/v2 v2.23.1
	github.com/davecgh/go-spew v1.1.1
	github.com/felixge/fgtrace v0.2.0
	github.com/google/go-cmp v0.7.0
	github.com/mna/pigeon v1.3.0
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.1
	github.com/pkg/errors v0.9.1
	github.com/pkg/profile v1.7.0
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/DataDog/gostackparse v0.6.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/felixge/fgprof v0.9.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.41.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// include support for disabling unexported fields
// TODO: still needed?
replace github.com/davecgh/go-spew => github.com/flw-cn/go-spew v1.1.2-0.20200624141737-10fccbfd0b23
