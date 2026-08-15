package output

import "fmt"

// CapabilityInfo is the user-facing help entry for a capability.
type CapabilityInfo struct {
	Name        string
	Description string
}

// PrintHelpWithCapabilities writes help text including available capabilities.
func (w *Writer) PrintHelpWithCapabilities(capabilities []CapabilityInfo) {
	fmt.Fprintf(w.Out, "Aryntra Aayam v%s\n\n", version)
	fmt.Fprintf(w.Out, "Project intelligence for developers.\n\n")

	fmt.Fprintf(w.Out, "Usage:\n")
	fmt.Fprintf(w.Out, "  Aryntra Aayam [target] [flags]\n")
	fmt.Fprintf(w.Out, "  Aryntra Aayam <capability> [target] [flags]\n\n")

	fmt.Fprintf(w.Out, "Flags:\n")
	fmt.Fprintf(w.Out, "  --help       Show this help message\n")
	fmt.Fprintf(w.Out, "  --version    Show version information\n")
	fmt.Fprintf(w.Out, "  --json       Output results as JSON\n")

	if len(capabilities) == 0 {
		return
	}

	fmt.Fprintf(w.Out, "\nCapabilities:\n")
	for _, c := range capabilities {
		fmt.Fprintf(w.Out, "  %-15s %s\n", c.Name, c.Description)
	}
}
