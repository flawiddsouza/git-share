package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X github.com/flawiddsouza/git-share/cmd.Version=x.y.z"
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of git-share",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("git-share %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
