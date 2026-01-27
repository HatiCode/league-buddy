package main

import (
	"context"
	"os"
	"time"

	"github.com/HatiCode/league-buddy/internal/riot"
	"github.com/HatiCode/league-buddy/internal/store"
	"github.com/HatiCode/league-buddy/pkg/ratelimit"
	"github.com/spf13/cobra"
)

var (
	apiKey   string
	platform string
	region   string // Derived from platform (americas, asia, europe, sea)
	dbURL    string

	riotClient *riot.APIClient
	dataStore  store.Store
)

var rootCmd = &cobra.Command{
	Use:   "league-buddy",
	Short: "League Buddy CLI - Get insights on your League of Legends gameplay",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if apiKey == "" {
			apiKey = os.Getenv("RIOT_API_KEY")
		}
		if apiKey == "" {
			cmd.PrintErrln("Error: RIOT_API_KEY is required (use --api-key or set RIOT_API_KEY env var)")
			os.Exit(1)
		}

		limiter := ratelimit.NewLimiter(
			ratelimit.WithLimit(500, 10*time.Second),
			ratelimit.WithLimit(30000, 10*time.Minute),
		)

		riotClient = riot.NewClient(apiKey, riot.WithRateLimiter(limiter))

		// Default region from platform if not specified
		if region == "" {
			if r, ok := riot.PlatformToRegion[platform]; ok {
				region = r
			} else {
				region = riot.RegionEurope // Fallback
			}
		}

		if dbURL == "" {
			dbURL = os.Getenv("DATABASE_URL")
		}
		if dbURL != "" {
			var err error
			dataStore, err = store.NewPostgresStore(context.Background(), dbURL)
			if err != nil {
				cmd.PrintErrf("Warning: failed to connect to database: %v\n", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Riot API key (or set RIOT_API_KEY env var)")
	rootCmd.PersistentFlags().StringVar(&platform, "platform", "euw1", "Platform for summoner data (euw1, na1, kr, etc.)")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "Region for account lookup (americas, asia, europe). Defaults based on platform.")
	rootCmd.PersistentFlags().StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL (or set DATABASE_URL env var)")
}
