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

var riotID string
var summonerSave bool

var summonerCmd = &cobra.Command{
	Use:   "summoner",
	Short: "Get summoner information",
	Long:  `Get summoner by Riot ID (gameName#tagLine), e.g., "Faker#KR1"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if riotID == "" {
			return fmt.Errorf("--riot-id is required (format: gameName#tagLine)")
		}

		parts := strings.SplitN(riotID, "#", 2)
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

		// Step 2: Get summoner by PUUID
		summonerStart := time.Now()
		summoner, err := riotClient.GetSummonerByPUUID(ctx, platform, account.PUUID)
		if err != nil {
			return fmt.Errorf("failed to get summoner: %w", err)
		}
		summonerDuration := time.Since(summonerStart)
		totalDuration := time.Since(start)

		if summonerSave && dataStore != nil {
			entity := store.SummonerFromAPI(account, summoner, platform)
			if err := dataStore.UpsertSummoner(ctx, entity); err != nil {
				return fmt.Errorf("failed to save summoner: %w", err)
			}
		}
		// Output combined info
		output := struct {
			Account  any    `json:"account"`
			Summoner any    `json:"summoner"`
			Timing   timing `json:"timing"`
		}{
			Account:  account,
			Summoner: summoner,
			Timing: timing{
				Account:  accountDuration.String(),
				Summoner: summonerDuration.String(),
				Total:    totalDuration.String(),
			},
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	},
}

type timing struct {
	Account  string `json:"account"`
	Summoner string `json:"summoner"`
	Total    string `json:"total"`
}

func init() {
	summonerCmd.Flags().StringVar(&riotID, "riot-id", "", "Riot ID (format: gameName#tagLine, e.g., Faker#KR1)")
	summonerCmd.Flags().BoolVar(&summonerSave, "save", false, "Save summoner data to DB")
	getCmd.AddCommand(summonerCmd)
}
