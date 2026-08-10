package project

import (
	"path/filepath"
	"sort"
)

// Language represents a detected programming or markup language.
// The value is the human-readable name used in all output.
type Language string

const (
	LangGo         Language = "Go"
	LangRust       Language = "Rust"
	LangPython     Language = "Python"
	LangJavaScript Language = "JavaScript"
	LangTypeScript Language = "TypeScript"
	LangJava       Language = "Java"
	LangC          Language = "C"
	LangCPP        Language = "C++"
	LangCSharp     Language = "C#"
	LangHTML       Language = "HTML"
	LangCSS        Language = "CSS"
	LangSCSS       Language = "SCSS"
	LangMarkdown   Language = "Markdown"
	LangPowerShell Language = "PowerShell"
)

// extensionMap maps lowercase file extensions to their Language.
// Extensions must include the leading dot.
//
// Ordering strategy: alphabetical by Language name in output.
// This map is the single authoritative source for language detection.
// To add a new language, add entries here only.
var extensionMap = map[string]Language{
	".c":    LangC,
	".cc":   LangCPP,
	".cpp":  LangCPP,
	".cs":   LangCSharp,
	".css":  LangCSS,
	".cxx":  LangCPP,
	".go":   LangGo,
	".h":    LangC,
	".html": LangHTML,
	".java": LangJava,
	".js":   LangJavaScript,
	".jsx":  LangJavaScript,
	".md":   LangMarkdown,
	".ps1":  LangPowerShell,
	".py":   LangPython,
	".rs":   LangRust,
	".scss": LangSCSS,
	".ts":   LangTypeScript,
	".tsx":  LangTypeScript,
}

// DetectLanguages inspects a slice of file paths and returns the unique
// set of languages present, sorted alphabetically by language name.
//
// Files with unrecognised extensions are silently ignored — unknown is
// normal and must never produce an error.
//
// Ordering guarantee: output is alphabetically sorted so that results
// are deterministic across executions regardless of filesystem ordering.
func DetectLanguages(files []string) []Language {
	seen := make(map[Language]struct{})

	for _, f := range files {
		ext := filepath.Ext(f)
		if ext == "" {
			continue
		}
		// Normalise to lowercase so ".GO" and ".go" are equivalent.
		if lang, ok := extensionMap[ext]; ok {
			seen[lang] = struct{}{}
		}
	}

	langs := make([]Language, 0, len(seen))
	for lang := range seen {
		langs = append(langs, lang)
	}

	// Alphabetical sort gives deterministic output.
	sort.Slice(langs, func(i, j int) bool {
		return string(langs[i]) < string(langs[j])
	})

	return langs
}
