package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// initRepo creates a temporary Git repository wired with a test identity.
// Returns the repository root path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("setup command %v failed: %v", args, err)
		}
	}
	return dir
}

// addCommit creates a file and commits it with the given message and author.
func addCommit(t *testing.T, dir, filename, message, author string) {
	t.Helper()

	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte("content"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cmds := [][]string{
		{"git", "add", filename},
		{"git", "-c", "user.name=" + author,
			"-c", "user.email=" + author + "@test.com",
			"commit", "-m", message},
	}
	for _, args := range cmds {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		if err := c.Run(); err != nil {
			t.Fatalf("commit command %v failed: %v", args, err)
		}
	}
}

// ---------------------------------------------------------------------------
// commitHistory tests
// ---------------------------------------------------------------------------

func TestCommitHistory_EmptyRepository(t *testing.T) {
	dir := initRepo(t)

	h, err := commitHistory(dir)
	if err != nil {
		t.Fatalf("commitHistory returned error on empty repo: %v", err)
	}
	if h.Count != 0 {
		t.Errorf("expected Count=0, got %d", h.Count)
	}
}

func TestCommitHistory_SingleCommit(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "file.txt", "initial commit", "Test User")

	h, err := commitHistory(dir)
	if err != nil {
		t.Fatalf("commitHistory: %v", err)
	}
	if h.Count != 1 {
		t.Errorf("expected Count=1, got %d", h.Count)
	}
	if h.Latest.Message != "initial commit" {
		t.Errorf("expected message %q, got %q", "initial commit", h.Latest.Message)
	}
	if h.Latest.Author != "Test User" {
		t.Errorf("expected author %q, got %q", "Test User", h.Latest.Author)
	}
	if h.Latest.Hash == "" {
		t.Error("expected non-empty hash")
	}
	if h.Latest.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	if h.Latest.Timestamp.Location() != time.UTC {
		t.Error("expected UTC timestamp")
	}
}

func TestCommitHistory_MultipleCommits(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "first", "Test User")
	addCommit(t, dir, "b.txt", "second", "Test User")
	addCommit(t, dir, "c.txt", "third", "Test User")

	h, err := commitHistory(dir)
	if err != nil {
		t.Fatalf("commitHistory: %v", err)
	}
	if h.Count != 3 {
		t.Errorf("expected Count=3, got %d", h.Count)
	}
	// Latest must be the most recent commit.
	if h.Latest.Message != "third" {
		t.Errorf("expected latest message %q, got %q", "third", h.Latest.Message)
	}
}

func TestCommitHistory_NonGitDirectory(t *testing.T) {
	dir := t.TempDir()
	// Not a Git repo -- commitHistory is an internal helper, but we can
	// test the public Discover which wraps it.
	info, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if info.IsRepository {
		t.Error("expected IsRepository=false for plain directory")
	}
	if info.History.Count != 0 {
		t.Error("expected zero history for non-repository")
	}
}

// ---------------------------------------------------------------------------
// repoContributors tests (M3.6)
// ---------------------------------------------------------------------------

func TestContributors_EmptyRepository(t *testing.T) {
	dir := initRepo(t)

	c, err := repoContributors(dir)
	if err != nil {
		t.Fatalf("repoContributors: %v", err)
	}
	if c.Count != 0 {
		t.Errorf("expected Count=0, got %d", c.Count)
	}
	if len(c.Items) != 0 {
		t.Errorf("expected empty Items, got %v", c.Items)
	}
}

func TestContributors_SingleContributor(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "commit one", "Alice")
	addCommit(t, dir, "b.txt", "commit two", "Alice")

	c, err := repoContributors(dir)
	if err != nil {
		t.Fatalf("repoContributors: %v", err)
	}
	if c.Count != 1 {
		t.Errorf("expected Count=1, got %d", c.Count)
	}
	if c.Items[0].Name != "Alice" {
		t.Errorf("expected name %q, got %q", "Alice", c.Items[0].Name)
	}
	if c.Items[0].CommitCount != 2 {
		t.Errorf("expected CommitCount=2, got %d", c.Items[0].CommitCount)
	}
}

func TestContributors_MultipleContributors(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "one", "Charlie")
	addCommit(t, dir, "b.txt", "two", "Alice")
	addCommit(t, dir, "c.txt", "three", "Charlie")
	addCommit(t, dir, "d.txt", "four", "Bob")
	addCommit(t, dir, "e.txt", "five", "Charlie")

	c, err := repoContributors(dir)
	if err != nil {
		t.Fatalf("repoContributors: %v", err)
	}
	if c.Count != 3 {
		t.Errorf("expected Count=3, got %d", c.Count)
	}
	// Ordering: count desc, then name asc for ties.
	if c.Items[0].Name != "Charlie" || c.Items[0].CommitCount != 3 {
		t.Errorf("expected Charlie/3 first, got %+v", c.Items[0])
	}
	if c.Items[1].Name != "Alice" || c.Items[1].CommitCount != 1 {
		t.Errorf("expected Alice/1 second, got %+v", c.Items[1])
	}
	if c.Items[2].Name != "Bob" || c.Items[2].CommitCount != 1 {
		t.Errorf("expected Bob/1 third, got %+v", c.Items[2])
	}
}

func TestContributors_DeterministicOrdering(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "one", "Zelda")
	addCommit(t, dir, "b.txt", "two", "Alice")
	addCommit(t, dir, "c.txt", "three", "Zelda")
	addCommit(t, dir, "d.txt", "four", "Alice")
	// Alice and Zelda both have 2 commits -- tie broken by name ascending.

	c, err := repoContributors(dir)
	if err != nil {
		t.Fatalf("repoContributors: %v", err)
	}
	if c.Items[0].Name != "Alice" {
		t.Errorf("expected Alice first in tie (name ascending), got %q", c.Items[0].Name)
	}
	if c.Items[1].Name != "Zelda" {
		t.Errorf("expected Zelda second in tie, got %q", c.Items[1].Name)
	}
}

// ---------------------------------------------------------------------------
// RepositoryHealth tests (M3.7)
// ---------------------------------------------------------------------------

func TestHealth_EmptyRepository(t *testing.T) {
	dir := initRepo(t)

	info, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if info.Health.HasCommits {
		t.Error("expected HasCommits=false for empty repo")
	}
	if info.Health.HasContributors {
		t.Error("expected HasContributors=false for empty repo")
	}
}

func TestHealth_PopulatedRepository(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "first commit", "Test User")

	info, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if !info.Health.HasCommits {
		t.Error("expected HasCommits=true")
	}
	if !info.Health.HasContributors {
		t.Error("expected HasContributors=true")
	}
	if !info.Health.WorkingTreeClean {
		t.Error("expected WorkingTreeClean=true after commit")
	}
	if info.Health.DetachedHEAD {
		t.Error("expected DetachedHEAD=false on normal branch")
	}
}

func TestHealth_DirtyWorkingTree(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "first", "Test User")

	// Add an untracked file to make the tree dirty.
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	info, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if info.Health.WorkingTreeClean {
		t.Error("expected WorkingTreeClean=false for dirty tree")
	}
}

// ---------------------------------------------------------------------------
// Full Discover integration (M3.8 readiness)
// ---------------------------------------------------------------------------

func TestDiscover_FullPipeline(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "main.go", "feat: initial implementation", "Dev One")
	addCommit(t, dir, "README.md", "docs: add readme", "Dev Two")
	addCommit(t, dir, "util.go", "refactor: extract utilities", "Dev One")

	info, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}

	if !info.IsRepository {
		t.Fatal("expected IsRepository=true")
	}
	if info.History.Count != 3 {
		t.Errorf("expected 3 commits, got %d", info.History.Count)
	}
	if info.History.Latest.Message != "refactor: extract utilities" {
		t.Errorf("unexpected latest message: %q", info.History.Latest.Message)
	}
	if info.Contributors.Count != 2 {
		t.Errorf("expected 2 contributors, got %d", info.Contributors.Count)
	}
	if !info.Health.HasCommits {
		t.Error("expected HasCommits=true")
	}
	if !info.Health.HasContributors {
		t.Error("expected HasContributors=true")
	}
}
