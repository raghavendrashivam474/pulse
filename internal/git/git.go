// Package git provides Git repository intelligence for Pulse.
//
// All Git operations are centralized here. No other package should
// invoke git commands directly. Callers receive typed results and
// Pulse-native errors -- never raw exec errors.
package git

import (
	"os/exec"
	"path/filepath"
	"strings"

	pulseErrors "pulse/internal/errors"
)

// WorkingTreeState describes whether the repository working tree is
// clean or has uncommitted changes.
type WorkingTreeState string

const (
	WorkingTreeClean   WorkingTreeState = "clean"
	WorkingTreeDirty   WorkingTreeState = "dirty"
	WorkingTreeUnknown WorkingTreeState = "unknown"
)

// GitInfo is the structured representation of a project's Git state.
type GitInfo struct {
	IsRepository     bool
	RepositoryRoot   string
	RepositoryName   string
	Branch           string
	WorkingTreeState WorkingTreeState
}

// Discover runs the full Git intelligence pipeline for the given
// directory and returns a GitInfo value.
func Discover(targetPath string) (GitInfo, error) {
	root, ok, err := repositoryRoot(targetPath)
	if err != nil {
		return GitInfo{}, err
	}
	if !ok {
		return GitInfo{IsRepository: false}, nil
	}

	name := repositoryName(root)

	branch, err := currentBranch(root)
	if err != nil {
		return GitInfo{}, err
	}

	state, err := workingTreeState(root)
	if err != nil {
		return GitInfo{}, err
	}

	return GitInfo{
		IsRepository:     true,
		RepositoryRoot:   root,
		RepositoryName:   name,
		Branch:           branch,
		WorkingTreeState: state,
	}, nil
}

// run executes a git command in dir and returns trimmed stdout.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", pulseErrors.Environment(
			"git command failed: git "+strings.Join(args, " "),
			err,
		)
	}
	return strings.TrimSpace(string(out)), nil
}

// repositoryRoot returns the Git repository root containing targetPath.
func repositoryRoot(targetPath string) (string, bool, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = targetPath
	out, err := cmd.Output()
	if err != nil {
		return "", false, nil
	}
	root := filepath.Clean(strings.TrimSpace(string(out)))
	return root, true, nil
}

// repositoryName derives the repository name from the root path.
func repositoryName(root string) string {
	return filepath.Base(root)
}

// currentBranch returns the active branch name, or "HEAD" for detached state.
func currentBranch(root string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "HEAD", nil
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "HEAD", nil
	}
	return branch, nil
}

// workingTreeState reports whether the working tree has uncommitted changes.
func workingTreeState(root string) (WorkingTreeState, error) {
	output, err := run(root, "status", "--porcelain")
	if err != nil {
		return WorkingTreeUnknown, err
	}
	if output == "" {
		return WorkingTreeClean, nil
	}
	return WorkingTreeDirty, nil
}
