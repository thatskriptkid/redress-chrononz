module github.com/goretk/redress

go 1.18

require (
	github.com/TcM1911/r2g2 v0.3.2
	github.com/cheynewallace/tabby v1.1.1
	github.com/goretk/gore v0.10.0
	github.com/spf13/cobra v1.2.1
)

require (
	github.com/google/go-querystring v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
)

require (
	github.com/google/go-github/v45 v45.2.0
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/arch v0.0.0-20220412001346-fc48f9fe4c15 // indirect
	golang.org/x/mod v0.5.1 // indirect
)

// This is used during development and disabled for release builds.
//replace github.com/goretk/gore => ./gore
