package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/flawiddsouza/git-share/internal/client"
	"github.com/flawiddsouza/git-share/internal/crypto"
	"github.com/flawiddsouza/git-share/internal/git"
)

var (
	receiveCommit bool
)

var receiveCmd = &cobra.Command{
	Use:   "receive <code>",
	Short: "Download, decrypt, and apply a git patch",
	Long: `Download an encrypted patch from the relay server, decrypt it
using the embedded passphrase, and apply it to the current repository.

The code is the full string output by the sender, e.g.:
  git-share receive k7Xm9pQ2wR-alpha-bravo-charlie-delta`,
	Args: cobra.MinimumNArgs(1),
	RunE: runReceive,
}

func init() {
	receiveCmd.Flags().BoolVar(&receiveCommit, "commit", false, "apply as a commit (cherry-pick style)")
	rootCmd.AddCommand(receiveCmd)
}

func runReceive(cmd *cobra.Command, args []string) error {
	// Support both "code" as single arg and "codeId word1-word2-word3-word4" as two args
	code := strings.Join(args, "-")

	// 1. Parse the combined code
	codeID, passphrase, err := crypto.ParseCode(code)
	if err != nil {
		return err
	}

	// 2. Make sure we're in a git repo
	_, err = git.FindRepoRoot()
	if err != nil {
		return err
	}

	// 3. Download from relay server
	fmt.Fprintf(os.Stderr, "Downloading patch...\n")
	c := client.New(serverURL)
	encodedData, err := c.Receive(codeID)
	if err != nil {
		return err
	}

	// 4. Decode base64
	encrypted, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("decoding data: %w", err)
	}

	// 5. Derive key and decrypt
	fmt.Fprintf(os.Stderr, "Decrypting...\n")
	key, err := crypto.DeriveKey(passphrase)
	if err != nil {
		return fmt.Errorf("deriving key: %w", err)
	}

	patch, err := crypto.Decrypt(encrypted, key)
	if err != nil {
		return err
	}

	// 6. Apply the patch
	fmt.Fprintf(os.Stderr, "Applying patch...\n")
	if err := git.ApplyPatch(patch, receiveCommit); err != nil {
		return err
	}

	// 7. Show stats
	stats, _ := git.PatchStats(patch)
	fmt.Fprintf(os.Stderr, "\nPatch applied successfully.\n")
	if stats != "" {
		fmt.Fprintf(os.Stderr, "\n%s\n", stats)
	}

	return nil
}
