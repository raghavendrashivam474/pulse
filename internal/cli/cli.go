// Package cli handles argument parsing and application flow for Pulse.
package cli

import (
	"fmt"
	"os"

	"pulse/internal/config"
	pulseErrors "pulse/internal/errors"
	"pulse/internal/output"
	"pulse/internal/project"
	"pulse/internal/scanner"
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
//
// Pipeline: Config -> Target -> Root -> Scan -> Detect -> Output
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

	// M2.1 — Validate the target path.
	target, err := project.ResolveTarget(cfg.TargetPath)
	if err != nil {
		w.PrintError(err.Error())
		return ExitFailure
	}

	// M2.2 — Discover the project root.
	rootResult := project.DiscoverRoot(target)

	// M2.3 — Scan the filesystem from the discovered root.
	inv, scanErr := scanner.Scan(rootResult.Root)
	if scanErr != nil {
		w.PrintError(fmt.Sprintf("filesystem scan failed: %s", scanErr.Error()))
		return ExitFailure
	}

	// M2.4 — Detect the project type.
	detection := project.DetectType(inv)

	if cfg.JSON {
		printJSONDiscovery(w, rootResult, inv, detection)
		return ExitSuccess
	}

	printDiscovery(w, rootResult, inv, detection)
	return ExitSuccess
}

// printDiscovery renders the S2 discovery result as human-readable text.
// M2.8 will own the polished output format; this is the functional pipeline result.
func printDiscovery(w *output.Writer, root project.RootDiscovery, inv scanner.Inventory, d project.Detection) {
	fmt.Fprintf(w.Out, "Pulse — Project Discovery\n\n")

	fmt.Fprintf(w.Out, "Project\n")
	fmt.Fprintf(w.Out, "  Root:   %s\n", root.Root)
	fmt.Fprintf(w.Out, "  Type:   %s\n", string(d.Primary))

	if !root.MarkerFound {
		fmt.Fprintf(w.Out, "  Note:   no project root marker found; using target as root\n")
	}

	if len(d.AllDetected) > 1 {
		fmt.Fprintf(w.Out, "  Also:   ")
		for i, t := range d.AllDetected[1:] {
			if i > 0 {
				fmt.Fprintf(w.Out, ", ")
			}
			fmt.Fprintf(w.Out, "%s", string(t))
		}
		fmt.Fprintf(w.Out, "\n")
	}

	fmt.Fprintf(w.Out, "\n")
	fmt.Fprintf(w.Out, "Filesystem\n")
	fmt.Fprintf(w.Out, "  Files:       %d\n", len(inv.Files))
	fmt.Fprintf(w.Out, "  Directories: %d\n", len(inv.Dirs))
}

// printJSONDiscovery renders minimal JSON discovery output.
// A structured JSON renderer will replace this in M2.8.
func printJSONDiscovery(w *output.Writer, root project.RootDiscovery, inv scanner.Inventory, d project.Detection) {
	fmt.Fprintf(w.Out, "{\n")
	fmt.Fprintf(w.Out, "  \"root\": %q,\n", root.Root)
	fmt.Fprintf(w.Out, "  \"type\": %q,\n", string(d.Primary))
	fmt.Fprintf(w.Out, "  \"files\": %d,\n", len(inv.Files))
	fmt.Fprintf(w.Out, "  \"directories\": %d\n", len(inv.Dirs))
	fmt.Fprintf(w.Out, "}\n")
}

// Main is the true application entry point, called from main.go.
func Main() {
	os.Exit(Run(os.Args[1:]))
}
