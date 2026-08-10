// Package cli handles argument parsing and application flow for Pulse.
package cli

import (
	"os"

	"pulse/internal/config"
	pulseErrors "pulse/internal/errors"
	"pulse/internal/output"
)

// ExitSuccess is the process exit code for a successful run.
const ExitSuccess = 0

// ExitFailure is the process exit code for any failure.
const ExitFailure = 1

// Args holds the parsed command-line arguments.
type Args struct {
	Help       bool
	Version    bool
	JSON       bool
	TargetPath string
}

// ParseArgs parses os.Args-style input into an Args struct.
func ParseArgs(args []string) (*Args, error) {
	parsed := &Args{}
	positional := 0

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help", "-h":
			parsed.Help = true
		case "--version", "-v":
			parsed.Version = true
		case "--json":
			parsed.JSON = true
		default:
			if len(args[i]) > 0 && args[i][0] == '-' {
				return nil, pulseErrors.User("unknown flag: " + args[i])
			}
			positional++
			if positional > 1 {
				return nil, pulseErrors.User("too many arguments: only one target path is allowed")
			}
			parsed.TargetPath = args[i]
		}
	}

	return parsed, nil
}

// Run is the main entry point for the Pulse CLI.
// It returns a process exit code.
func Run(args []string) int {
	w := output.Default()

	parsed, err := ParseArgs(args)
	if err != nil {
		w.PrintError(err.Error())
		return ExitFailure
	}

	if parsed.Help {
		w.PrintHelp()
		return ExitSuccess
	}

	if parsed.Version {
		w.PrintVersion()
		return ExitSuccess
	}

	cfg, err := config.New(parsed.TargetPath, parsed.JSON)
	if err != nil {
		w.PrintError("could not resolve target path: " + err.Error())
		return ExitFailure
	}

	if cfg.JSON {
		if err := w.PrintJSON(); err != nil {
			w.PrintError("failed to produce JSON output: " + err.Error())
			return ExitFailure
		}
		return ExitSuccess
	}

	w.PrintSummary()
	return ExitSuccess
}

// Main is the true application entry point, called from main.go.
func Main() {
	os.Exit(Run(os.Args[1:]))
}
