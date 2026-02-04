package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/HatiCode/league-buddy/internal/analysis"
	"github.com/HatiCode/league-buddy/internal/coaching"
	"github.com/HatiCode/league-buddy/internal/models"
	"github.com/spf13/cobra"
)

var (
	coachRiotID      string
	coachMatchCount  int
	coachProvider    string
	coachLLMKey      string
	coachModel       string
	coachMaxTokens   int64
	coachTemperature float64
)

var coachCmd = &cobra.Command{
	Use:   "coach",
	Short: "Get AI coaching advice based on match analysis",
	Long:  `Analyze recent matches and get personalized coaching advice from an AI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if coachRiotID == "" {
			return fmt.Errorf("--riot-id is required (format: gameName#tagLine)")
		}

		parts := strings.SplitN(coachRiotID, "#", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid Riot ID format, expected gameName#tagLine")
		}
		gameName, tagLine := parts[0], parts[1]

		ctx := context.Background()
		start := time.Now()

		account, err := riotClient.GetAccountByRiotID(ctx, region, gameName, tagLine)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		entries, err := riotClient.GetLeagueEntries(ctx, platform, account.PUUID)
		if err != nil {
			return fmt.Errorf("failed to get league entries: %w", err)
		}
		var soloEntry *models.LeagueEntry
		for i := range entries {
			if entries[i].QueueType == models.QueueRankedSolo {
				soloEntry = &entries[i]
				break
			}
		}

		var previousMatchIDs map[string]bool
		if dataStore != nil {
			prevSession, err := dataStore.GetLatestCoachingSession(ctx, account.PUUID)
			if err != nil {
				return fmt.Errorf("failed to get previous session: %w", err)
			}
			if prevSession != nil {
				var ids []string
				if err := json.Unmarshal(prevSession.MatchIDs, &ids); err == nil {
					previousMatchIDs = make(map[string]bool, len(ids))
					for _, id := range ids {
						previousMatchIDs[id] = true
					}
				}
			}
		}

		allMatchIDs, err := riotClient.GetMatchIDs(ctx, platform, account.PUUID, coachMatchCount, models.QueueIDRankedSolo)
		if err != nil {
			return fmt.Errorf("failed to get match IDs: %w", err)
		}
		if len(allMatchIDs) == 0 {
			return fmt.Errorf("no matches found for this summoner")
		}

		var matchIDs []string
		for _, id := range allMatchIDs {
			if !previousMatchIDs[id] {
				matchIDs = append(matchIDs, id)
			}
		}
		if len(matchIDs) == 0 {
			cmd.Println("No new matches since last coaching session.")
			return nil
		}

		var matches []models.Match
		timelines := make(map[string]*models.Timeline)
		for _, id := range matchIDs {
			match, err := riotClient.GetMatch(ctx, platform, id)
			if err != nil {
				cmd.PrintErrf("Warning: failed to fetch match %s: %v\n", id, err)
				continue
			}
			matches = append(matches, *match)

			tl, err := riotClient.GetMatchTimeline(ctx, platform, id)
			if err == nil {
				timelines[id] = tl
			}
		}
		if len(matches) == 0 {
			return fmt.Errorf("failed to fetch any match details")
		}
		riotDuration := time.Since(start)

		analysisStart := time.Now()
		playerAnalysis, err := analysis.AnalyzePlayer(analysis.PlayerAnalysisParams{
			PUUID:     account.PUUID,
			GameName:  account.GameName,
			TagLine:   account.TagLine,
			Matches:   matches,
			Timelines: timelines,
			League:    soloEntry,
		})
		if err != nil {
			return fmt.Errorf("failed to analyze matches: %w", err)
		}
		analysisDuration := time.Since(analysisStart)

		llmClient, err := createLLMClient()
		if err != nil {
			return err
		}

		coachingStart := time.Now()
		svc := coaching.NewService(llmClient, dataStore)
		resp, err := svc.Coach(ctx, playerAnalysis, matchIDs)
		if err != nil {
			return fmt.Errorf("coaching failed: %w", err)
		}
		coachingDuration := time.Since(coachingStart)
		totalDuration := time.Since(start)

		output := struct {
			Player   coachPlayerInfo            `json:"player"`
			Coaching *coaching.CoachingResponse `json:"coaching"`
			Timing   coachTimingInfo            `json:"timing"`
		}{
			Player: coachPlayerInfo{
				RiotID:  fmt.Sprintf("%s#%s", account.GameName, account.TagLine),
				Tier:    playerAnalysis.Tier,
				Rank:    playerAnalysis.Rank,
				WinRate: playerAnalysis.WinRate,
			},
			Coaching: resp,
			Timing: coachTimingInfo{
				RiotAPI:  riotDuration.String(),
				Analysis: analysisDuration.String(),
				Coaching: coachingDuration.String(),
				Total:    totalDuration.String(),
			},
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	},
}

type coachPlayerInfo struct {
	RiotID  string  `json:"riotId"`
	Tier    string  `json:"tier,omitempty"`
	Rank    string  `json:"rank,omitempty"`
	WinRate float64 `json:"winRate"`
}

type coachTimingInfo struct {
	RiotAPI  string `json:"riotApi"`
	Analysis string `json:"analysis"`
	Coaching string `json:"coaching"`
	Total    string `json:"total"`
}

func createLLMClient() (coaching.LLMClient, error) {
	key := coachLLMKey
	if key == "" {
		switch coachProvider {
		case "claude":
			key = os.Getenv("ANTHROPIC_API_KEY")
		case "openai":
			key = os.Getenv("OPENAI_API_KEY")
		}
	}
	if key == "" {
		return nil, fmt.Errorf("LLM API key is required (use --llm-key or set ANTHROPIC_API_KEY/OPENAI_API_KEY env var)")
	}

	switch coachProvider {
	case "claude":
		return coaching.NewClaudeClient(coaching.ClaudeConfig{
			APIKey:      key,
			Model:       coachModel,
			MaxTokens:   coachMaxTokens,
			Temperature: coachTemperature,
		})
	case "openai":
		return coaching.NewOpenAIClient(coaching.OpenAIConfig{
			APIKey:      key,
			Model:       coachModel,
			MaxTokens:   coachMaxTokens,
			Temperature: coachTemperature,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %q (use claude or openai)", coachProvider)
	}
}

func init() {
	coachCmd.Flags().StringVar(&coachRiotID, "riot-id", "", "Riot ID (format: gameName#tagLine, e.g., Faker#KR1)")
	coachCmd.Flags().IntVar(&coachMatchCount, "match-count", 10, "Number of recent matches to analyze")
	coachCmd.Flags().StringVar(&coachProvider, "provider", "claude", "LLM provider (claude, openai)")
	coachCmd.Flags().StringVar(&coachLLMKey, "llm-key", "", "LLM API key (or set ANTHROPIC_API_KEY/OPENAI_API_KEY env var)")
	coachCmd.Flags().StringVar(&coachModel, "model", "", "LLM model (defaults based on provider)")
	coachCmd.Flags().Int64Var(&coachMaxTokens, "max-tokens", 0, "Max response tokens (default: provider default)")
	coachCmd.Flags().Float64Var(&coachTemperature, "temperature", 0, "LLM temperature (default: provider default)")
	rootCmd.AddCommand(coachCmd)
}
