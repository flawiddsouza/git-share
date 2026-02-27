package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// FindRepoRoot returns the root directory of the current git repository.
func FindRepoRoot() (string, error) {
	out, err := runGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not a git repository (or any parent): %w", err)
	}
	return strings.TrimSpace(out), nil
}

// GetDiff returns the diff of uncommitted changes in the working tree.
func GetDiff() ([]byte, error) {
	out, err := runGit("diff", "--binary")
	if err != nil {
		return nil, fmt.Errorf("getting diff: %w", err)
	}
	if out == "" {
		stagedOut, _ := runGit("diff", "--cached", "--name-only")
		if stagedOut != "" {
			return nil, errors.New("no uncommitted changes found (did you mean to use 'git-share --staged'?)")
		}
		return nil, errors.New("no uncommitted changes found")
	}
	return []byte(out), nil
}

// GetStagedDiff returns the diff of staged changes.
func GetStagedDiff() ([]byte, error) {
	out, err := runGit("diff", "--cached", "--binary")
	if err != nil {
		return nil, fmt.Errorf("getting staged diff: %w", err)
	}
	if out == "" {
		unstagedOut, _ := runGit("diff", "--name-only")
		if unstagedOut != "" {
			return nil, errors.New("no staged changes found (did you mean to use 'git-share'?)")
		}
		return nil, errors.New("no staged changes found")
	}
	return []byte(out), nil
}

// GetCommitPatch returns the patch for a commit or commit range using format-patch.
// Accepts: single SHA, branch name, HEAD~3.., commit1..commit2, etc.
func GetCommitPatch(commitRef string) ([]byte, error) {
	var out string
	var err error

	// If it looks like a range (contains ".."), use it directly
	if strings.Contains(commitRef, "..") {
		out, err = runGit("format-patch", "--stdout", commitRef)
	} else {
		// Single ref — verify it's a valid commit first
		_, verifyErr := runGit("cat-file", "-t", commitRef)
		if verifyErr != nil {
			return nil, fmt.Errorf("invalid commit reference %q (not found or not a commit)", commitRef)
		}
		// Use -1 to get exactly that one commit as a patch
		out, err = runGit("format-patch", "--stdout", "-1", commitRef)
	}

	if err != nil {
		return nil, fmt.Errorf("getting commit patch for %q: %w", commitRef, err)
	}
	if out == "" {
		return nil, fmt.Errorf("no commits found for %q", commitRef)
	}
	return []byte(out), nil
}

// ApplyPatch applies a patch to the current repository.
// If forceAm is true, it uses `git am` to create a commit.
// Otherwise, it uses `git apply` to only update the working tree/index.
func ApplyPatch(patch []byte, forceAm bool) error {
	if forceAm {
		// Use git am to create a commit (cherry-pick style)
		err := runGitWithStdin(patch, "am")
		if err != nil {
			// Abort any failed am
			_ = runGitWithStdin(nil, "am", "--abort")
			return fmt.Errorf("failed to apply commit via 'git am': %w", err)
		}
		return nil
	}

	// Use git apply (works for both simple diffs and format-patch output, but only applies changes)
	err := runGitWithStdin(patch, "apply")
	if err != nil {
		return fmt.Errorf("failed to apply patch via 'git apply': %w", err)
	}

	return nil
}

// PatchStats returns a human-readable summary of what a patch would change.
func PatchStats(patch []byte) (string, error) {
	out, err := runGitWithStdinOutput(patch, "apply", "--stat")
	if err != nil {
		// Try diffstat format for format-patch output
		out, err = runGitWithStdinOutput(patch, "apply", "--stat", "--check")
		if err != nil {
			return "", nil // silently ignore, stats are optional
		}
	}
	return strings.TrimRight(out, "\r\n "), nil
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", errMsg)
	}
	return stdout.String(), nil
}

func runGitWithStdin(stdin []byte, args ...string) error {
	cmd := exec.Command("git", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("%s", errMsg)
	}
	return nil
}

func runGitWithStdinOutput(stdin []byte, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", errMsg)
	}
	return stdout.String(), nil
}
