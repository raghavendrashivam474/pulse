package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pulse/internal/git"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "test@pulse.test")
	mustGit(t, dir, "config", "user.name", "Pulse Test")
	return dir
}

func initRepoWithCommit(t *testing.T) string {
	t.Helper()
	dir := initRepo(t)
	writeFile(t, dir, "README.md", "# test")
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "initial commit")
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("writeFile: %v", err)
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// M3.1 -- Repository detection

func TestDiscover_NotARepository(t *testing.T) {
	dir := t.TempDir()
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.IsRepository {
		t.Error("expected IsRepository=false for plain directory")
	}
}

func TestDiscover_IsRepository(t *testing.T) {
	dir := initRepoWithCommit(t)
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.IsRepository {
		t.Error("expected IsRepository=true for git repository")
	}
}

func TestDiscover_NestedTarget_InsideRepository(t *testing.T) {
	root := initRepoWithCommit(t)
	sub := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	info, err := git.Discover(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.IsRepository {
		t.Error("expected IsRepository=true for nested directory inside repo")
	}
	if info.RepositoryRoot != root {
		t.Errorf("RepositoryRoot = %q; want %q", info.RepositoryRoot, root)
	}
}

// M3.2 -- Repository metadata

func TestDiscover_RepositoryRoot(t *testing.T) {
	dir := initRepoWithCommit(t)
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.RepositoryRoot != dir {
		t.Errorf("RepositoryRoot = %q; want %q", info.RepositoryRoot, dir)
	}
}

func TestDiscover_RepositoryName(t *testing.T) {
	dir := initRepoWithCommit(t)
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Base(dir)
	if info.RepositoryName != want {
		t.Errorf("RepositoryName = %q; want %q", info.RepositoryName, want)
	}
}

func TestDiscover_TargetDistinctFromRepositoryRoot(t *testing.T) {
	root := initRepoWithCommit(t)
	sub := filepath.Join(root, "subdir")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	info, err := git.Discover(sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub == info.RepositoryRoot {
		t.Error("analysis target and repository root should differ for nested target")
	}
	if info.RepositoryRoot != root {
		t.Errorf("RepositoryRoot = %q; want %q", info.RepositoryRoot, root)
	}
}

// M3.3 -- Branch / HEAD state

func TestDiscover_Branch(t *testing.T) {
	dir := initRepoWithCommit(t)
	mustGit(t, dir, "checkout", "-b", "main")
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Branch != "main" {
		t.Errorf("Branch = %q; want %q", info.Branch, "main")
	}
}

func TestDiscover_DetachedHEAD(t *testing.T) {
	dir := initRepoWithCommit(t)
	out, err := exec.Command("git", "-C", dir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	hash := strings.TrimSpace(string(out))
	mustGit(t, dir, "checkout", hash)
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Branch != "HEAD" {
		t.Errorf("Branch = %q; want HEAD for detached HEAD", info.Branch)
	}
}

// M3.4 -- Working-tree state

func TestDiscover_WorkingTreeClean(t *testing.T) {
	dir := initRepoWithCommit(t)
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.WorkingTreeState != git.WorkingTreeClean {
		t.Errorf("WorkingTreeState = %q; want clean", info.WorkingTreeState)
	}
}

func TestDiscover_WorkingTreeDirty(t *testing.T) {
	dir := initRepoWithCommit(t)
	writeFile(t, dir, "dirty.txt", "uncommitted change")
	info, err := git.Discover(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.WorkingTreeState != git.WorkingTreeDirty {
		t.Errorf("WorkingTreeState = %q; want dirty", info.WorkingTreeState)
	}
}
