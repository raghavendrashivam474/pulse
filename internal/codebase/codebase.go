// Package codebase provides the structured model of a project's source
// artifacts, packages, and their relationships.
//
// This package is the foundation for code structure intelligence.
// It represents what exists in the codebase (files, directories, packages)
// and how those things relate (dependencies).
//
// Design constraints:
//   - All collections are deterministically ordered.
//   - No output-specific formatting; this is a domain model.
//   - No filesystem handles or transient runtime state.
//   - Language-specific discovery (e.g. Go packages) is isolated
//     behind clear interfaces so future languages can be added.
//
// Architecture:
//
//	Filesystem / Go source
//	        │
//	        ▼
//	 Codebase Model
//	        │
//	        ├── Terminal
//	        ├── JSON
//	        └── Future Graph
package codebase

import (
	"path/filepath"
	"sort"
	"strings"

	"pulse/internal/project"
	"pulse/internal/scanner"
)

// File represents a single source artifact in the codebase.
//
// Future graph role: Node.
type File struct {
	// Path is the forward-slash-separated path relative to the project root.
	Path string

	// Name is the filename without directory components.
	Name string

	// Extension is the lowercase file extension including the leading dot.
	// Empty string for files with no extension (e.g. Makefile).
	Extension string

	// Language is the detected language for this file, or empty if unknown.
	Language project.Language
}

// Directory represents a filesystem container in the codebase.
//
// Future graph role: Node.
type Directory struct {
	// Path is the forward-slash-separated path relative to the project root.
	Path string

	// Name is the directory name without parent components.
	Name string
}

// BuildFiles converts scanner file entries into codebase File domain objects.
//
// The result is sorted alphabetically by Path for deterministic ordering.
// Language is assigned via the project.extensionMap through DetectLanguages
// lookup logic — we use the same extension-to-language mapping.
func BuildFiles(entries []scanner.FileEntry) []File {
	files := make([]File, len(entries))
	for i, e := range entries {
		files[i] = File{
			Path:      e.RelPath,
			Name:      fileName(e.RelPath),
			Extension: e.Extension,
			Language:  languageForExtension(e.Extension),
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return files
}

// BuildDirectories converts scanner directory entries into codebase
// Directory domain objects.
//
// The result is sorted alphabetically by Path for deterministic ordering.
func BuildDirectories(entries []scanner.DirEntry) []Directory {
	dirs := make([]Directory, len(entries))
	for i, e := range entries {
		dirs[i] = Directory{
			Path: e.RelPath,
			Name: dirName(e.RelPath),
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	return dirs
}

// fileName extracts the final component of a forward-slash path.
func fileName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// dirName extracts the final component of a forward-slash directory path.
func dirName(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

// extensionLanguageMap mirrors project.extensionMap for file-level language
// assignment. We maintain a local copy to avoid exporting internal maps
// from the project package.
//
// This is intentionally kept in sync with project.extensionMap.
// If a language is added there, add it here as well.
var extensionLanguageMap = map[string]project.Language{
	".c":    project.LangC,
	".cc":   project.LangCPP,
	".cpp":  project.LangCPP,
	".cs":   project.LangCSharp,
	".css":  project.LangCSS,
	".cxx":  project.LangCPP,
	".go":   project.LangGo,
	".h":    project.LangC,
	".html": project.LangHTML,
	".java": project.LangJava,
	".js":   project.LangJavaScript,
	".jsx":  project.LangJavaScript,
	".md":   project.LangMarkdown,
	".ps1":  project.LangPowerShell,
	".py":   project.LangPython,
	".rs":   project.LangRust,
	".scss": project.LangSCSS,
	".ts":   project.LangTypeScript,
	".tsx":  project.LangTypeScript,
}

// languageForExtension returns the Language for a given file extension,
// or empty string if unrecognised.
func languageForExtension(ext string) project.Language {
	ext = strings.ToLower(ext)
	if lang, ok := extensionLanguageMap[ext]; ok {
		return lang
	}
	return ""
}

// LanguageForExtension is the exported version for use by other packages
// that need file-level language assignment.
func LanguageForExtension(ext string) project.Language {
	return languageForExtension(ext)
}

// FilesByDirectory groups files by their parent directory path.
// The key is the directory path (forward-slash separated, relative to root).
// Files at the root level use "" as the key.
func FilesByDirectory(files []File) map[string][]File {
	result := make(map[string][]File)
	for _, f := range files {
		dir := filepath.ToSlash(filepath.Dir(f.Path))
		if dir == "." {
			dir = ""
		}
		result[dir] = append(result[dir], f)
	}
	return result
}
