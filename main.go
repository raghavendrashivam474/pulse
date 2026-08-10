package main

import "pulse/internal/cli"

// version is set at build time via:
//
//	go build -ldflags "-X main.version=1.0.0"
//
// During normal development builds it remains the default below.
var version = "1.0.0"

func main() {
	_ = version
	cli.Main()
}
