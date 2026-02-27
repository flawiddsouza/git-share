package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/flawiddsouza/git-share/internal/client"
	"github.com/flawiddsouza/git-share/internal/crypto"
	"github.com/flawiddsouza/git-share/internal/git"
)

var (
	SendStaged bool
	SendTTL    string
)

var sendCmd = &cobra.Command{
	Use:   "send [commit or range]",
	Short: "Encrypt and upload git changes to the relay server",
	Long: `Collect git changes, encrypt them with a one-time passphrase,
and upload to the relay server. Outputs a code for the receiver.

Examples:
  git-share send                       # uncommitted working tree changes
  git-share send --staged              # staged changes only
  git-share send abc123                # a specific commit (by SHA)
  git-share send HEAD~3..              # last 3 commits
  git-share send main..feature         # commits in feature not in main`,
	RunE: RunSend,
}

func init() {
	sendCmd.Flags().BoolVar(&SendStaged, "staged", false, "send staged changes only")
	sendCmd.Flags().StringVar(&SendTTL, "ttl", "1h", "time-to-live for the patch (e.g. 15m, 1h)")
	rootCmd.AddCommand(sendCmd)
}

type sendDeps interface {
	FindRepoRoot() (string, error)
	GetCommitPatch(ref string) ([]byte, error)
	GetStagedDiff() ([]byte, error)
	GetDiff() ([]byte, error)
	GenerateCode() (code, codeID, passphrase string, err error)
	DeriveKey(passphrase string) ([]byte, error)
	Encrypt(data, key []byte) ([]byte, error)
	Send(codeID, data string, ttl int) (*client.SendResponse, error)
}

type realSendDeps struct{}

func (d realSendDeps) FindRepoRoot() (string, error) { return git.FindRepoRoot() }
func (d realSendDeps) GetCommitPatch(ref string) ([]byte, error) {
	return git.GetCommitPatch(ref)
}
func (d realSendDeps) GetStagedDiff() ([]byte, error) { return git.GetStagedDiff() }
func (d realSendDeps) GetDiff() ([]byte, error)       { return git.GetDiff() }
func (d realSendDeps) GenerateCode() (string, string, string, error) {
	return crypto.GenerateCode()
}
func (d realSendDeps) DeriveKey(passphrase string) ([]byte, error) {
	return crypto.DeriveKey(passphrase)
}
func (d realSendDeps) Encrypt(data, key []byte) ([]byte, error) {
	return crypto.Encrypt(data, key)
}
func (d realSendDeps) Send(codeID, data string, ttl int) (*client.SendResponse, error) {
	c := client.New(serverURL)
	return c.Send(codeID, data, ttl)
}

func RunSend(cmd *cobra.Command, args []string) error {
	return runSendWithDeps(os.Stdout, os.Stderr, realSendDeps{}, args, SendStaged, SendTTL)
}

func runSendWithDeps(stdout, stderr interface {
	Write([]byte) (int, error)
}, deps sendDeps, args []string, staged bool, ttlStr string) error {
	// 1. Make sure we're in a git repo
	_, err := deps.FindRepoRoot()
	if err != nil {
		return err
	}

	// 2. Collect the patch
	fmt.Fprintf(stderr, "Collecting changes...\n")
	var patch []byte
	isCommit := false

	switch {
	case len(args) > 0:
		// Positional arg = commit ref or range
		patch, err = deps.GetCommitPatch(args[0])
		isCommit = true
	case staged:
		patch, err = deps.GetStagedDiff()
	default:
		patch, err = deps.GetDiff()
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(stderr, "   Found %d bytes of changes\n", len(patch))

	// 3. Generate the code (codeID + passphrase)
	code, codeID, passphrase, err := deps.GenerateCode()
	if err != nil {
		return fmt.Errorf("generating code: %w", err)
	}

	// 4. Derive encryption key and encrypt
	key, err := deps.DeriveKey(passphrase)
	if err != nil {
		return fmt.Errorf("deriving key: %w", err)
	}

	encrypted, err := deps.Encrypt(patch, key)
	if err != nil {
		return fmt.Errorf("encrypting: %w", err)
	}

	// 5. Parse TTL
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return fmt.Errorf("invalid TTL %q: %w", ttlStr, err)
	}

	// 6. Upload to relay server
	fmt.Fprintf(stderr, "Encrypting and uploading...\n")
	encoded := base64.StdEncoding.EncodeToString(encrypted)

	resp, err := deps.Send(codeID, encoded, int(ttl.Seconds()))
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// 7. Print the receive command
	fmt.Fprintf(stderr, "\nEncrypted and uploaded.\n")
	fmt.Fprintf(stderr, "Share this with the receiver:\n\n")
	fmt.Fprintf(stdout, "   git-share receive %s\n", code)
	if isCommit {
		fmt.Fprintf(stderr, "OR to receive as a commit instead of a patch:\n")
		fmt.Fprintf(stdout, "   git-share receive %s --commit\n", code)
	}
	fmt.Fprintf(stderr, "\nExpires: %s | One-time use only\n", resp.Expiry)

	return nil
}
