package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/flawiddsouza/git-share/internal/server"
)

var (
	servePort   int
	serveMaxTTL string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the relay server",
	Long: `Start the git-share relay server. The server stores encrypted blobs
in memory and serves them once before deleting. Blobs expire after the
configured TTL.

This can be self-hosted or used as a public relay.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 3141, "port to listen on")
	serveCmd.Flags().StringVar(&serveMaxTTL, "max-ttl", "1h", "maximum TTL for stored patches")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	maxTTL, err := time.ParseDuration(serveMaxTTL)
	if err != nil {
		return fmt.Errorf("invalid max-ttl %q: %w", serveMaxTTL, err)
	}

	config := server.DefaultConfig()
	config.Port = servePort
	config.MaxTTL = maxTTL

	srv := server.New(config)
	return srv.Start()
}
