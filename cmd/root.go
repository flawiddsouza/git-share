package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	defaultServer = "https://git-share.artelin.dev"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "git-share",
	Short: "Securely share git patches with E2E encryption",
	Long: `git-share is a CLI tool for sharing git patches securely.

It encrypts your changes with a one-time passphrase, uploads the encrypted
blob to a relay server, and gives you a code to share. The receiver uses
the code to download, decrypt, and apply the patch. The patch is destroyed
after a single use.

Think of it as "croc" but specifically for git patches.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", defaultServer, "relay server URL")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
