package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HatiCode/league-buddy/internal/store"
	"github.com/spf13/cobra"
)

var matchRiotID string
var matchSave bool

var matchCmd = &cobra.Command{
	Use:   "match",
	Short: "Get latest match for a summoner",
	Long:  `Get the most recent match for a summoner by Riot ID (gameName#tagLine)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if matchRiotID == "" {
			return fmt.Errorf("--riot-id is required (format: gameName#tagLine)")
		}

		parts := strings.SplitN(matchRiotID, "#", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid Riot ID format, expected gameName#tagLine")
		}
		gameName, tagLine := parts[0], parts[1]

		ctx := context.Background()
		start := time.Now()

		// Step 1: Get account by Riot ID
		account, err := riotClient.GetAccountByRiotID(ctx, region, gameName, tagLine)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}
		accountDuration := time.Since(start)

		// Step 2: Get latest match ID
		matchIDsStart := time.Now()
		matchIDs, err := riotClient.GetMatchIDs(ctx, platform, account.PUUID, 1)
		if err != nil {
			return fmt.Errorf("failed to get match IDs: %w", err)
		}
		if len(matchIDs) == 0 {
			return fmt.Errorf("no matches found for this summoner")
		}
		matchIDsDuration := time.Since(matchIDsStart)

		// Step 3: Get match details
		matchStart := time.Now()
		match, err := riotClient.GetMatch(ctx, platform, matchIDs[0])
		if err != nil {
			return fmt.Errorf("failed to get match: %w", err)
		}
		matchDuration := time.Since(matchStart)
		totalDuration := time.Since(start)

		// Save to DB if enabled
		if matchSave && dataStore != nil {
			matchEntity := store.MatchFromAPI(match)
			participants := store.ParticipantsFromAPI(match)
			if err := dataStore.SaveMatch(ctx, matchEntity, participants); err != nil {
				return fmt.Errorf("failed to save match: %w", err)
			}
		}

		// Output combined info
		output := struct {
			MatchID string `json:"matchId"`
			Match   any    `json:"match"`
			Timing  struct {
				Account  string `json:"account"`
				MatchIDs string `json:"matchIds"`
				Match    string `json:"match"`
				Total    string `json:"total"`
			} `json:"timing"`
		}{
			MatchID: matchIDs[0],
			Match:   match,
		}
		output.Timing.Account = accountDuration.String()
		output.Timing.MatchIDs = matchIDsDuration.String()
		output.Timing.Match = matchDuration.String()
		output.Timing.Total = totalDuration.String()

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	},
}

func init() {
	matchCmd.Flags().StringVar(&matchRiotID, "riot-id", "", "Riot ID (format: gameName#tagLine, e.g., Faker#KR1)")
	matchCmd.Flags().BoolVar(&matchSave, "save", false, "Save match data to DB")
	getCmd.AddCommand(matchCmd)
}
