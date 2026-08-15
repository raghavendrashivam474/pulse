package cli

import (
	"os"

	"github.com/raghavendrashivam474/aayam/internal/capability"
	"github.com/raghavendrashivam474/aayam/internal/config"
	aayamErrors "github.com/raghavendrashivam474/aayam/internal/errors"
	"github.com/raghavendrashivam474/aayam/internal/output"
	"github.com/raghavendrashivam474/aayam/internal/snapshot"
)

const ExitSuccess = 0
const ExitFailure = 1

// Args holds parsed CLI arguments.
type Args struct {
	Help           bool
	Version        bool
	JSON           bool
	CapabilityName string
	TargetPath     string
	Positionals    []string
}

// ParseArgs parses os.Args-style input into an Args struct.
func ParseArgs(args []string) (*Args, error) {
	parsed := &Args{}

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
				return nil, aayamErrors.User("unknown flag: " + args[i])
			}

			parsed.Positionals = append(parsed.Positionals, args[i])
			if len(parsed.Positionals) > 2 {
				return nil, aayamErrors.User("too many arguments: expected [capability] [target] or [target]")
			}
		}
	}

	if len(parsed.Positionals) == 1 {
		parsed.TargetPath = parsed.Positionals[0]
	}

	return parsed, nil
}

// ResolveCommand maps positional arguments into either legacy target mode
// or capability mode.
func ResolveCommand(parsed *Args, registry *capability.Registry) error {
	switch len(parsed.Positionals) {
	case 0:
		parsed.CapabilityName = ""
		parsed.TargetPath = ""
		return nil

	case 1:
		first := parsed.Positionals[0]
		if registry != nil && registry.IsKnown(first) {
			parsed.CapabilityName = first
			parsed.TargetPath = "."
			return nil
		}

		parsed.CapabilityName = ""
		parsed.TargetPath = first
		return nil

	case 2:
		first := parsed.Positionals[0]
		second := parsed.Positionals[1]

		if registry == nil || !registry.IsKnown(first) {
			return aayamErrors.User("unknown capability: " + first)
		}

		parsed.CapabilityName = first
		parsed.TargetPath = second
		return nil

	default:
		return aayamErrors.User("too many arguments: expected [capability] [target] or [target]")
	}
}

// Run executes the CLI and returns a process exit code.
func Run(args []string) int {
	w := output.Default()
	registry := newCapabilityRegistry(w)

	parsed, err := ParseArgs(args)
	if err != nil {
		w.PrintError(err.Error())
		return ExitFailure
	}

	if parsed.Help {
		w.PrintHelpWithCapabilities(capabilityHelpEntries(registry))
		return ExitSuccess
	}

	if parsed.Version {
		w.PrintVersion()
		return ExitSuccess
	}

	if err := ResolveCommand(parsed, registry); err != nil {
		w.PrintError(err.Error())
		return ExitFailure
	}

	// No target and no capability selected: preserve existing default behavior.
	if parsed.TargetPath == "" && parsed.CapabilityName == "" {
		if parsed.JSON {
			if jsonErr := w.PrintJSON(); jsonErr != nil {
				w.PrintError(jsonErr.Error())
				return ExitFailure
			}
			return ExitSuccess
		}

		w.PrintSummary()
		return ExitSuccess
	}

	cfg, err := config.New(parsed.TargetPath, parsed.JSON)
	if err != nil {
		w.PrintError("could not resolve target path: " + err.Error())
		return ExitFailure
	}

	snap, err := snapshot.Discover(cfg.TargetPath)
	if err != nil {
		w.PrintError(err.Error())
		return ExitFailure
	}

	if parsed.CapabilityName != "" {
		c, ok := registry.Lookup(parsed.CapabilityName)
		if !ok {
			w.PrintError("unknown capability: " + parsed.CapabilityName)
			return ExitFailure
		}

		if _, err := c.Run(capability.Context{
			Snap: snap,
			JSON: cfg.JSON,
		}); err != nil {
			w.PrintError(err.Error())
			return ExitFailure
		}

		return ExitSuccess
	}

	if cfg.JSON {
		if jsonErr := w.PrintDiscoveryJSON(snap); jsonErr != nil {
			w.PrintError(jsonErr.Error())
			return ExitFailure
		}
		return ExitSuccess
	}

	w.PrintDiscovery(snap)
	return ExitSuccess
}

// Main is the true application entry point, called from main.go.
func Main() {
	os.Exit(Run(os.Args[1:]))
}
