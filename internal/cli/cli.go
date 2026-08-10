package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/raghavendrashivam474/pulse/internal/output"
)

const version = "v1.0.0"

// App is the Pulse CLI application.
type App struct {
	stdout io.Writer
	stderr io.Writer
}

// New creates a new App with default output streams.
func New() *App {
	return &App{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Options holds the parsed CLI arguments.
type Options struct {
	Path        string
	JSON        bool
	ShowVersion bool
	ShowHelp    bool
}

// Run parses arguments and executes the appropriate action.
func (a *App) Run(args []string) error {
	fs := flag.NewFlagSet("pulse", flag.ContinueOnError)
	fs.SetOutput(a.stderr)

	var opts Options

	fs.BoolVar(&opts.JSON, "json", false, "Output machine-readable JSON")
	fs.BoolVar(&opts.ShowVersion, "version", false, "Show version")
	fs.BoolVar(&opts.ShowHelp, "help", false, "Show help")

	fs.Usage = func() {
		fmt.Fprintf(a.stdout, "Pulse %s - project intelligence for developers\n\n", version)
		fmt.Fprintf(a.stdout, "Usage:\n")
		fmt.Fprintf(a.stdout, "    pulse [path] [options]\n\n")
		fmt.Fprintf(a.stdout, "Options:\n")
		fmt.Fprintf(a.stdout, "    --help       Show help\n")
		fmt.Fprintf(a.stdout, "    --version    Show version\n")
		fmt.Fprintf(a.stdout, "    --json       Output machine-readable JSON\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if opts.ShowHelp {
		fs.Usage()
		return nil
	}

	if opts.ShowVersion {
		fmt.Fprintf(a.stdout, "Pulse %s\n", version)
		return nil
	}

	// Capture optional positional path argument.
	if fs.NArg() > 0 {
		opts.Path = fs.Arg(0)
	}

	renderer := output.NewRenderer(a.stdout, opts.JSON)
	renderer.RenderDefault()

	return nil
}
