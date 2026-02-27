package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing and returns its path
// and a cleanup function.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	// Create temp dir
	dir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Helper to run commands in the temp dir
	runCmd := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run %s %v: %v", name, args, err)
		}
	}

	// Initialize git repo
	runCmd("git", "init")

	// Set user config for commits
	runCmd("git", "config", "user.email", "test@example.com")
	runCmd("git", "config", "user.name", "Test User")

	// Create initial commit so we have a HEAD
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("initial\n"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runCmd("git", "add", "test.txt")
	runCmd("git", "commit", "-m", "initial commit")

	// Save original working directory to restore later
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Change to the temp repo directory
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Failed to chdir to temp dir: %v", err)
	}

	cleanup := func() {
		os.Chdir(originalWd)
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestFindRepoRoot(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Root level
	root, err := FindRepoRoot()
	if err != nil {
		t.Errorf("FindRepoRoot failed at root: %v", err)
	}
	evalRoot, _ := filepath.EvalSymlinks(root)
	evalDir, _ := filepath.EvalSymlinks(dir)
	if !strings.EqualFold(evalRoot, evalDir) {
		t.Errorf("Expected root %q, got %q", dir, root)
	}

	// 2. Subdirectory
	subDir := filepath.Join(dir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Failed to chdir to subdir: %v", err)
	}
	root, err = FindRepoRoot()
	if err != nil {
		t.Errorf("FindRepoRoot failed in subdir: %v", err)
	}
	evalRoot, _ = filepath.EvalSymlinks(root)
	if !strings.EqualFold(evalRoot, evalDir) {
		t.Errorf("Expected root %q from subdir, got %q", dir, root)
	}

	// 3. Not a git repo
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	_, err = FindRepoRoot()
	if err == nil {
		t.Error("Expected error for non-git directory, got nil")
	}
}

func TestGetDiff(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Clean working directory
	_, err := GetDiff()
	if err == nil {
		t.Error("Expected error for clean working directory, got nil")
	} else if err.Error() != "no uncommitted changes found" {
		t.Errorf("Expected 'no uncommitted changes found', got %q", err.Error())
	}

	// 2. Unstaged changes only
	if err := os.WriteFile("test.txt", []byte("unstaged\n"), 0644); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	diff, err := GetDiff()
	if err != nil {
		t.Errorf("Expected nil error for unstaged changes, got %v", err)
	}
	if !bytes.Contains(diff, []byte("-initial")) || !bytes.Contains(diff, []byte("+unstaged")) {
		t.Errorf("Diff does not contain expected changes: %s", diff)
	}

	// 3. Staged changes only hint
	exec.Command("git", "add", "test.txt").Run()
	_, err = GetDiff()
	if err == nil {
		t.Error("Expected error for staged changes only, got nil")
	} else if !strings.Contains(err.Error(), "did you mean to use 'git-share --staged'?") {
		t.Errorf("Expected hint for staged changes, got %q", err.Error())
	}

	// 4. Binary file
	binData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	if err := os.WriteFile("binary.bin", binData, 0644); err != nil {
		t.Fatalf("Failed to write binary file: %v", err)
	}
	// Git needs binary files to be tracked to show in diff usually
	exec.Command("git", "add", "binary.bin").Run()
	exec.Command("git", "commit", "-m", "add binary").Run()
	if err := os.WriteFile("binary.bin", append(binData, 0xAA), 0644); err != nil {
		t.Fatalf("Failed to modify binary file: %v", err)
	}
	diff, err = GetDiff()
	if err != nil {
		t.Errorf("Failed to get binary diff: %v", err)
	}
	if !bytes.Contains(diff, []byte("GIT binary patch")) {
		t.Errorf("Diff does not reflect binary change: %s", diff)
	}
}

func TestGetStagedDiff(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Clean working directory
	_, err := GetStagedDiff()
	if err == nil {
		t.Error("Expected error for clean working directory, got nil")
	} else if err.Error() != "no staged changes found" {
		t.Errorf("Expected 'no staged changes found', got %q", err.Error())
	}

	// 2. Staged changes only
	if err := os.WriteFile("test.txt", []byte("staged\n"), 0644); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	exec.Command("git", "add", "test.txt").Run()
	diff, err := GetStagedDiff()
	if err != nil {
		t.Errorf("Expected nil error for staged changes, got %v", err)
	}
	if !bytes.Contains(diff, []byte("-initial")) || !bytes.Contains(diff, []byte("+staged")) {
		t.Errorf("Diff does not contain expected changes: %s", diff)
	}

	// 3. Unstaged changes only hint
	exec.Command("git", "reset", "--hard", "HEAD").Run()
	if err := os.WriteFile("test.txt", []byte("unstaged\n"), 0644); err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}
	_, err = GetStagedDiff()
	if err == nil {
		t.Error("Expected error for unstaged changes only, got nil")
	} else if !strings.Contains(err.Error(), "did you mean to use 'git-share'?") {
		t.Errorf("Expected hint for unstaged changes, got %q", err.Error())
	}

	// 4. Rename and Deletion
	exec.Command("git", "reset", "--hard", "HEAD").Run()
	if err := os.Rename("test.txt", "renamed.txt"); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
	exec.Command("git", "add", "renamed.txt").Run()
	exec.Command("git", "rm", "test.txt").Run()
	diff, err = GetStagedDiff()
	if err != nil {
		t.Errorf("Staged diff for rename/delete failed: %v", err)
	}
	if !bytes.Contains(diff, []byte("rename from test.txt")) || !bytes.Contains(diff, []byte("rename to renamed.txt")) {
		t.Logf("Git version might not show as rename if content not committed yet. Checking basic rm/add...")
		if !bytes.Contains(diff, []byte("deleted file mode")) {
			t.Errorf("Diff missing deletion info: %s", diff)
		}
	}
}

func TestGetCommitPatch(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create another commit
	if err := os.WriteFile("test.txt", []byte("v2\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "second commit").Run()

	// 1. Test single commit (HEAD)
	patch, err := GetCommitPatch("HEAD")
	if err != nil {
		t.Errorf("GetCommitPatch(HEAD) failed: %v", err)
	}
	if !bytes.Contains(patch, []byte("Subject: [PATCH] second commit")) {
		t.Errorf("Patch missing subject: %s", patch)
	}

	// 2. Test range (HEAD~1..)
	patch, err = GetCommitPatch("HEAD~1..")
	if err != nil {
		t.Errorf("GetCommitPatch(HEAD~1..) failed: %v", err)
	}
	if !bytes.Contains(patch, []byte("Subject: [PATCH] second commit")) {
		t.Errorf("Range patch missing expected commit: %s", patch)
	}

	// 3. Test invalid ref
	_, err = GetCommitPatch("nonexistent-ref")
	if err == nil {
		t.Errorf("Expected error for invalid ref, got nil")
	}
}

func TestApplyPatchStrategies(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Test standard 'git apply' via GetDiff output
	if err := os.WriteFile("test.txt", []byte("modified\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	diff, err := GetDiff()
	if err != nil {
		t.Fatalf("Failed to get diff: %v", err)
	}
	exec.Command("git", "checkout", "test.txt").Run()
	if err := ApplyPatch(diff, false); err != nil {
		t.Errorf("ApplyPatch (simple) failed: %v", err)
	}
	content, _ := os.ReadFile("test.txt")
	if string(content) != "modified\n" {
		t.Errorf("Simple patch apply verification failed: %s", content)
	}

	// 2. Test 'git am' fallback via GetCommitPatch output
	// Create another commit first
	if err := os.WriteFile("second.txt", []byte("second\n"), 0644); err != nil {
		t.Fatalf("Failed to write second file: %v", err)
	}
	exec.Command("git", "add", "second.txt").Run()
	exec.Command("git", "commit", "-m", "second commit").Run()

	patch, _ := GetCommitPatch("HEAD")
	// Undo the commit to test applying it back
	exec.Command("git", "reset", "--hard", "HEAD~1").Run()
	if err := ApplyPatch(patch, false); err != nil {
		t.Errorf("ApplyPatch (apply) failed: %v", err)
	}
	// Verify file exists now
	if _, err := os.Stat("second.txt"); os.IsNotExist(err) {
		t.Errorf("ApplyPatch (am/apply fallback) did not restore file")
	}

	// 3. Binary patch
	exec.Command("git", "reset", "--hard", "HEAD").Run()
	binData := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	os.WriteFile("bin", binData, 0644)
	exec.Command("git", "add", "bin").Run()
	exec.Command("git", "commit", "-m", "add bin").Run()
	os.WriteFile("bin", append(binData, 0x00), 0644)
	binDiff, _ := GetDiff()
	exec.Command("git", "checkout", "bin").Run()
	if err := ApplyPatch(binDiff, false); err != nil {
		t.Errorf("Binary ApplyPatch failed: %v", err)
	}
	content, _ = os.ReadFile("bin")
	if len(content) != 5 {
		t.Errorf("Binary patch apply verification failed")
	}
}

func TestPatchStats(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	if err := os.WriteFile("test.txt", []byte("stats\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	diff, _ := GetDiff()
	stats, err := PatchStats(diff)
	if err != nil {
		t.Errorf("PatchStats failed: %v", err)
	}
	if !strings.Contains(stats, "test.txt") {
		t.Errorf("Stats output unexpected: %s", stats)
	}
}

func TestGetCommitPatchRange(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create 3 additional commits
	for i := 1; i <= 3; i++ {
		fname := fmt.Sprintf("file%d.txt", i)
		if err := os.WriteFile(fname, []byte(fmt.Sprintf("content %d\n", i)), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", fname, err)
		}
		if err := exec.Command("git", "add", fname).Run(); err != nil {
			t.Fatalf("Failed to git add %s: %v", fname, err)
		}
		if err := exec.Command("git", "commit", "-m", fmt.Sprintf("commit %d", i)).Run(); err != nil {
			t.Fatalf("Failed to commit %d: %v", i, err)
		}
	}

	// Get patch for last 2 commits (commit 2 and commit 3)
	patch, err := GetCommitPatch("HEAD~2..")
	if err != nil {
		t.Fatalf("Failed to get range patch: %v", err)
	}

	// Verify both commit subjects are in the stdout stream
	// Note: format-patch uses [PATCH 1/2] etc. for ranges
	if !strings.Contains(string(patch), "Subject: [PATCH 1/2] commit 2") && !strings.Contains(string(patch), "Subject: [PATCH] commit 2") {
		t.Errorf("Patch missing 'commit 2' or unexpected format. Patch snippet: %s", patch)
	}
	if !strings.Contains(string(patch), "Subject: [PATCH 2/2] commit 3") && !strings.Contains(string(patch), "Subject: [PATCH] commit 3") {
		t.Errorf("Patch missing 'commit 3' or unexpected format. Patch snippet: %s", patch)
	}
}

func TestSpecialFilenames(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	fname := "file with spaces.txt"
	content := []byte("special content\n")
	if err := os.WriteFile(fname, content, 0644); err != nil {
		t.Fatalf("Failed to write special file: %v", err)
	}

	// 1. Test Diff
	if err := exec.Command("git", "add", fname).Run(); err != nil {
		t.Fatalf("Failed to git add special file: %v", err)
	}
	diff, err := GetStagedDiff()
	if err != nil {
		t.Fatalf("Failed to get diff for special filename: %v", err)
	}
	if !bytes.Contains(diff, []byte("file with spaces.txt")) {
		t.Errorf("Diff missing special filename: %s", diff)
	}

	// 2. Test Apply
	exec.Command("git", "reset", "--hard", "HEAD").Run()
	if _, err := os.Stat(fname); err == nil {
		t.Fatalf("File should be gone after reset")
	}

	if err := ApplyPatch(diff, false); err != nil {
		t.Fatalf("Failed to apply patch with special filename: %v", err)
	}
	if _, err := os.Stat(fname); err != nil {
		t.Errorf("File with spaces not restored: %v", err)
	}
}

func TestApplyConflict(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Initial change
	os.WriteFile("test.txt", []byte("version A\n"), 0644)
	diff, _ := GetDiff()

	// Diverge the file
	os.WriteFile("test.txt", []byte("version B\n"), 0644)
	exec.Command("git", "add", "test.txt").Run()
	exec.Command("git", "commit", "-m", "diverged").Run()

	// Attempt to apply the "version A" patch
	err := ApplyPatch(diff, false)
	if err == nil {
		t.Error("Expected conflict error, got nil")
	}
}

func TestTagSupport(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	exec.Command("git", "tag", "mytag").Run()
	patch, err := GetCommitPatch("mytag")
	if err != nil {
		t.Errorf("Failed to use tag as ref: %v", err)
	}
	if !bytes.Contains(patch, []byte("Subject: [PATCH] initial commit")) {
		t.Errorf("Tag patch missing expected content: %s", patch)
	}
}

func TestApplyPatchCommit(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// 1. Create a commit to send
	if err := os.WriteFile("commit_file.txt", []byte("commit content\n"), 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}
	exec.Command("git", "add", "commit_file.txt").Run()
	exec.Command("git", "commit", "-m", "explicit commit message").Run()

	patch, err := GetCommitPatch("HEAD")
	if err != nil {
		t.Fatalf("Failed to get patch: %v", err)
	}

	// 2. Reset to previous state
	exec.Command("git", "reset", "--hard", "HEAD~1").Run()
	if _, err := os.Stat("commit_file.txt"); err == nil {
		t.Fatalf("File should be gone after reset")
	}

	// 3. Apply with forceAm=true
	if err := ApplyPatch(patch, true); err != nil {
		t.Fatalf("ApplyPatch(forceAm=true) failed: %v", err)
	}

	// 4. Verify commit exists
	out, err := exec.Command("git", "log", "-1", "--pretty=%s").Output()
	if err != nil {
		t.Fatalf("Failed to run git log: %v", err)
	}
	if strings.TrimSpace(string(out)) != "explicit commit message" {
		t.Errorf("Expected commit message 'explicit commit message', got %q", string(out))
	}
	if _, err := os.Stat("commit_file.txt"); err != nil {
		t.Errorf("File not restored after commit apply: %v", err)
	}
}
