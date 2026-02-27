package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/flawiddsouza/git-share/internal/client"
)

type mockSendDeps struct {
	repoRoot    string
	patch       []byte
	err         error
	code        string
	codeID      string
	passphrase  string
	expiry      string
	capturedRef string
}

func (m *mockSendDeps) FindRepoRoot() (string, error) { return m.repoRoot, nil }
func (m *mockSendDeps) GetCommitPatch(ref string) ([]byte, error) {
	m.capturedRef = ref
	return m.patch, m.err
}
func (m *mockSendDeps) GetStagedDiff() ([]byte, error) { return m.patch, m.err }
func (m *mockSendDeps) GetDiff() ([]byte, error)       { return m.patch, m.err }
func (m *mockSendDeps) GenerateCode() (string, string, string, error) {
	return m.code, m.codeID, m.passphrase, nil
}
func (m *mockSendDeps) DeriveKey(passphrase string) ([]byte, error) { return []byte("key"), nil }
func (m *mockSendDeps) Encrypt(data, key []byte) ([]byte, error)    { return data, nil }
func (m *mockSendDeps) Send(codeID, data string, ttl int) (*client.SendResponse, error) {
	return &client.SendResponse{Expiry: m.expiry}, nil
}

func TestRunSendWithDeps(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		staged        bool
		patch         string
		wantStdout    []string
		wantStderr    []string
		notWantStdout []string
		notWantStderr []string
	}{
		{
			name:          "uncommitted changes",
			args:          []string{},
			patch:         "diff content",
			wantStdout:    []string{"git-share receive abc-123"},
			notWantStdout: []string{"--commit"},
			notWantStderr: []string{"OR to receive as a commit instead of a patch:"},
		},
		{
			name:          "staged changes",
			args:          []string{},
			staged:        true,
			patch:         "diff content",
			wantStdout:    []string{"git-share receive abc-123"},
			notWantStdout: []string{"--commit"},
			notWantStderr: []string{"OR to receive as a commit instead of a patch:"},
		},
		{
			name:       "specific commit",
			args:       []string{"HEAD"},
			patch:      "patch content",
			wantStdout: []string{"git-share receive abc-123", "git-share receive abc-123 --commit"},
			wantStderr: []string{"OR to receive as a commit instead of a patch:"},
		},
		{
			name:       "commit range",
			args:       []string{"main..feature"},
			patch:      "patch content",
			wantStdout: []string{"git-share receive abc-123", "git-share receive abc-123 --commit"},
			wantStderr: []string{"OR to receive as a commit instead of a patch:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			deps := &mockSendDeps{
				repoRoot:   "/repo",
				patch:      []byte(tt.patch),
				code:       "abc-123",
				codeID:     "id",
				passphrase: "pass",
				expiry:     "2026-02-27T17:00:00Z",
			}

			err := runSendWithDeps(stdout, stderr, deps, tt.args, tt.staged, "1h")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			stdoutStr := stdout.String()
			stderrStr := stderr.String()

			for _, want := range tt.wantStdout {
				if !strings.Contains(stdoutStr, want) {
					t.Errorf("stdout missing %q\nGOT:\n%s", want, stdoutStr)
				}
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(stderrStr, want) {
					t.Errorf("stderr missing %q\nGOT:\n%s", want, stderrStr)
				}
			}
			for _, notWant := range tt.notWantStdout {
				if strings.Contains(stdoutStr, notWant) {
					t.Errorf("stdout should NOT contain %q\nGOT:\n%s", notWant, stdoutStr)
				}
			}
			for _, notWant := range tt.notWantStderr {
				if strings.Contains(stderrStr, notWant) {
					t.Errorf("stderr should NOT contain %q\nGOT:\n%s", notWant, stderrStr)
				}
			}
		})
	}
}
