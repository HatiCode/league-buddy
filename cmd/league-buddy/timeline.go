package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var timelineRiotID string
var timelineMatchID string

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Get match timeline data",
	Long:  `Get timeline data for a match (frame-by-frame game state)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		start := time.Now()

		var matchID string

		// If match ID provided directly, use it
		if timelineMatchID != "" {
			matchID = timelineMatchID
		} else if timelineRiotID != "" {
			// Otherwise, get latest match for riot ID
			parts := strings.SplitN(timelineRiotID, "#", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid Riot ID format, expected gameName#tagLine")
			}
			gameName, tagLine := parts[0], parts[1]

			account, err := riotClient.GetAccountByRiotID(ctx, region, gameName, tagLine)
			if err != nil {
				return fmt.Errorf("failed to get account: %w", err)
			}

			matchIDs, err := riotClient.GetMatchIDs(ctx, platform, account.PUUID, 1)
			if err != nil {
				return fmt.Errorf("failed to get match IDs: %w", err)
			}
			if len(matchIDs) == 0 {
				return fmt.Errorf("no matches found for this summoner")
			}
			matchID = matchIDs[0]
		} else {
			return fmt.Errorf("either --match-id or --riot-id is required")
		}

		// Get timeline
		timelineStart := time.Now()
		timeline, err := riotClient.GetMatchTimeline(ctx, platform, matchID)
		if err != nil {
			return fmt.Errorf("failed to get timeline: %w", err)
		}
		timelineDuration := time.Since(timelineStart)
		totalDuration := time.Since(start)

		// Output
		output := struct {
			MatchID  string `json:"matchId"`
			Timeline any    `json:"timeline"`
			Timing   struct {
				Timeline string `json:"timeline"`
				Total    string `json:"total"`
			} `json:"timing"`
		}{
			MatchID:  matchID,
			Timeline: timeline,
		}
		output.Timing.Timeline = timelineDuration.String()
		output.Timing.Total = totalDuration.String()

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	},
}

func init() {
	timelineCmd.Flags().StringVar(&timelineRiotID, "riot-id", "", "Riot ID to get latest match timeline (format: gameName#tagLine)")
	timelineCmd.Flags().StringVar(&timelineMatchID, "match-id", "", "Match ID to get timeline for (e.g., EUW1_1234567890)")
	getCmd.AddCommand(timelineCmd)
}
