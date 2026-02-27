package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/flawiddsouza/git-share/internal/server"
)

var (
	servePort    int
	serveMaxTTL  string
	serveMaxSize string
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
	serveCmd.Flags().StringVar(&serveMaxSize, "max-size", "10MB", "maximum blob size (e.g. 5MB, 512KB, 1GB)")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	maxTTL, err := time.ParseDuration(serveMaxTTL)
	if err != nil {
		return fmt.Errorf("invalid max-ttl %q: %w", serveMaxTTL, err)
	}

	maxSize, err := parseByteSize(serveMaxSize)
	if err != nil {
		return fmt.Errorf("invalid max-size %q: %w", serveMaxSize, err)
	}

	config := server.DefaultConfig()
	config.Port = servePort
	config.MaxTTL = maxTTL
	config.MaxSize = maxSize

	srv := server.New(config)
	return srv.Start()
}

// parseByteSize parses a human-readable byte size string like "10MB", "512KB", "1GB".
func parseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Find where the numeric part ends
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}
	if i == 0 {
		return 0, fmt.Errorf("no numeric value found")
	}

	numStr := s[:i]
	unit := strings.TrimSpace(strings.ToUpper(s[i:]))

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", numStr, err)
	}

	var multiplier float64
	switch unit {
	case "B", "":
		multiplier = 1
	case "KB", "K":
		multiplier = 1024
	case "MB", "M":
		multiplier = 1024 * 1024
	case "GB", "G":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown unit %q (use B, KB, MB, or GB)", unit)
	}

	result := int64(num * multiplier)
	if result <= 0 {
		return 0, fmt.Errorf("size must be greater than zero")
	}

	return result, nil
}
