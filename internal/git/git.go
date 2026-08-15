// Package git provides Git repository intelligence for Aryntra Aayam.
//
// All Git operations are centralized here. No other package should
// invoke git commands directly. Callers receive typed results and
// Aryntra Aayam-native errors -- never raw exec errors.
package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	aayamErrors "github.com/raghavendrashivam474/aayam/internal/errors"
)

// WorkingTreeState describes whether the repository working tree is
// clean or has uncommitted changes.
type WorkingTreeState string

const (
	WorkingTreeClean   WorkingTreeState = "clean"
	WorkingTreeDirty   WorkingTreeState = "dirty"
	WorkingTreeUnknown WorkingTreeState = "unknown"
)

// CommitInfo holds structured information about a single Git commit.
type CommitInfo struct {
	Hash      string
	Message   string
	Author    string
	Timestamp time.Time
}

// CommitHistory holds aggregate commit history for the repository.
type CommitHistory struct {
	Count  int
	Latest CommitInfo
}

// Contributor represents a single contributor and their commit count.
type Contributor struct {
	Name        string
	CommitCount int
}

// Contributors holds the full contributor picture for the repository.
type Contributors struct {
	Count int
	Items []Contributor
}

// RepositoryHealth holds objective, deterministic repository state signals.
// There is no score or subjective interpretation here -- only facts.
type RepositoryHealth struct {
	HasCommits       bool
	HasContributors  bool
	WorkingTreeClean bool
	DetachedHEAD     bool
}

// GitInfo is the structured representation of a project's Git state.
type GitInfo struct {
	IsRepository     bool
	RepositoryRoot   string
	RepositoryName   string
	Branch           string
	WorkingTreeState WorkingTreeState
	History          CommitHistory
	Contributors     Contributors
	Health           RepositoryHealth
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

	history, err := commitHistory(root)
	if err != nil {
		return GitInfo{}, err
	}

	contributors, err := repoContributors(root)
	if err != nil {
		return GitInfo{}, err
	}

	detached := branch == "HEAD"

	health := RepositoryHealth{
		HasCommits:       history.Count > 0,
		HasContributors:  contributors.Count > 0,
		WorkingTreeClean: state == WorkingTreeClean,
		DetachedHEAD:     detached,
	}

	return GitInfo{
		IsRepository:     true,
		RepositoryRoot:   root,
		RepositoryName:   name,
		Branch:           branch,
		WorkingTreeState: state,
		History:          history,
		Contributors:     contributors,
		Health:           health,
	}, nil
}

// run executes a git command in dir and returns trimmed stdout.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", aayamErrors.Environment(
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

// commitHistory returns the commit count and latest commit for the repository.
// An empty repository (no commits) is a valid state and returns a zero value.
func commitHistory(root string) (CommitHistory, error) {
	// Count total commits. This fails gracefully on an empty repository.
	countOut, err := run(root, "rev-list", "--count", "HEAD")
	if err != nil {
		// Empty repository: no commits yet. This is not an error.
		return CommitHistory{Count: 0}, nil
	}

	count, err := strconv.Atoi(countOut)
	if err != nil {
		return CommitHistory{}, aayamErrors.Environment(
			fmt.Sprintf("could not parse commit count %q", countOut),
			err,
		)
	}

	if count == 0 {
		return CommitHistory{Count: 0}, nil
	}

	// Use a delimiter that will never appear in a commit message.
	// Format: hash<US>subject<US>author name<US>unix timestamp
	// US = ASCII unit separator (0x1F).
	const sep = "\x1f"
	format := fmt.Sprintf("--format=%%H%s%%s%s%%aN%s%%ct", sep, sep, sep)

	logOut, err := run(root, "log", "-1", format)
	if err != nil {
		return CommitHistory{}, err
	}

	parts := strings.Split(logOut, sep)
	if len(parts) != 4 {
		return CommitHistory{}, aayamErrors.Environment(
			fmt.Sprintf("unexpected git log output: %q", logOut),
			nil,
		)
	}

	unixSec, err := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
	if err != nil {
		return CommitHistory{}, aayamErrors.Environment(
			fmt.Sprintf("could not parse commit timestamp %q", parts[3]),
			err,
		)
	}

	latest := CommitInfo{
		Hash:      strings.TrimSpace(parts[0]),
		Message:   strings.TrimSpace(parts[1]),
		Author:    strings.TrimSpace(parts[2]),
		Timestamp: time.Unix(unixSec, 0).UTC(),
	}

	return CommitHistory{
		Count:  count,
		Latest: latest,
	}, nil
}

// repoContributors returns all contributors ordered by commit count
// descending, then name ascending. An empty repository returns a
// zero-value Contributors.
func repoContributors(root string) (Contributors, error) {
	// git shortlog aggregates commit counts per author name.
	// -s  = summary (count only, no commit subjects)
	// -n  = sort by count descending
	// -e is intentionally omitted: we do not expose email addresses.
	// HEAD must exist; on an empty repository this returns exit code 128.
	cmd := exec.Command("git", "shortlog", "-sn", "HEAD")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		// Empty repository -- no commits, no contributors.
		return Contributors{Count: 0}, nil
	}

	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return Contributors{Count: 0}, nil
	}

	lines := strings.Split(raw, "\n")
	items := make([]Contributor, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Each line: "   <count>\t<name>"
		tabIdx := strings.Index(line, "\t")
		if tabIdx < 0 {
			continue
		}
		countStr := strings.TrimSpace(line[:tabIdx])
		name := strings.TrimSpace(line[tabIdx+1:])

		n, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}
		items = append(items, Contributor{Name: name, CommitCount: n})
	}

	// git shortlog -sn already sorts by count descending.
	// For ties we need a secondary name-ascending sort so output is
	// fully deterministic across Git versions and platforms.
	stableSort(items)

	return Contributors{
		Count: len(items),
		Items: items,
	}, nil
}

// stableSort sorts contributors by commit count descending, then name ascending.
// Implemented without importing sort to keep the dependency surface minimal.
func stableSort(items []Contributor) {
	// Insertion sort is fine here -- contributor lists are small.
	for i := 1; i < len(items); i++ {
		for j := i; j > 0; j-- {
			a, b := items[j-1], items[j]
			if a.CommitCount < b.CommitCount ||
				(a.CommitCount == b.CommitCount && a.Name > b.Name) {
				items[j-1], items[j] = items[j], items[j-1]
			} else {
				break
			}
		}
	}
}
