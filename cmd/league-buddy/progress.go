package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/HatiCode/league-buddy/internal/coaching"
	"github.com/guptarohit/asciigraph"
	"github.com/spf13/cobra"
)

var (
	progressRiotID string
	progressGraph  bool
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Show coaching progress trend over time",
	Long:  `Load all coaching sessions for a player and output trend data as JSON or ASCII graphs. Requires a database connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if progressRiotID == "" {
			return fmt.Errorf("--riot-id is required (format: gameName#tagLine)")
		}

		parts := strings.SplitN(progressRiotID, "#", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid Riot ID format, expected gameName#tagLine")
		}
		gameName, tagLine := parts[0], parts[1]

		if dataStore == nil {
			return fmt.Errorf("database is required for progress tracking (use --db-url or set DATABASE_URL)")
		}

		ctx := context.Background()

		account, err := riotClient.GetAccountByRiotID(ctx, region, gameName, tagLine)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}

		svc := coaching.NewService(nil, dataStore)
		progress, err := svc.GetProgress(ctx, account.PUUID)
		if err != nil {
			return fmt.Errorf("failed to get progress: %w", err)
		}

		if progress.Sessions == 0 {
			cmd.Println("No coaching sessions found. Run 'league-buddy coach' first.")
			return nil
		}

		if progressGraph {
			renderGraphs(progress)
			return nil
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(progress)
	},
}

func renderGraphs(progress *coaching.PlayerProgress) {
	fmt.Printf("Progress: %s#%s (%d sessions)\n\n", progress.GameName, progress.TagLine, progress.Sessions)

	fmt.Print("Rank:  ")
	for i, tp := range progress.Trend {
		if i > 0 {
			fmt.Print(" -> ")
		}
		label := tp.Tier + " " + tp.Rank
		fmt.Printf("%s (%s)", label, tp.SessionDate.Format("Jan 02"))
	}
	fmt.Println()

	type metricDef struct {
		name   string
		values []float64
		format string
	}

	metrics := []metricDef{
		{name: "Win Rate (%)", format: "%.0f"},
		{name: "KDA", format: "%.1f"},
		{name: "CS/min", format: "%.1f"},
		{name: "Vision/min", format: "%.2f"},
		{name: "Deaths/min", format: "%.2f"},
		{name: "Gold/min", format: "%.0f"},
	}

	for i := range metrics {
		metrics[i].values = make([]float64, len(progress.Trend))
	}

	for i, tp := range progress.Trend {
		metrics[0].values[i] = tp.WinRate * 100
		metrics[1].values[i] = tp.Averages.KDA
		metrics[2].values[i] = tp.Averages.CSPerMinute
		metrics[3].values[i] = tp.Averages.VisionScorePerMinute
		metrics[4].values[i] = tp.Averages.DeathsPerMinute
		metrics[5].values[i] = tp.Averages.GoldPerMinute
	}

	for _, m := range metrics {
		fmt.Println(m.name)
		fmt.Println(asciigraph.Plot(m.values,
			asciigraph.Height(8),
			asciigraph.Width(40),
			asciigraph.Precision(2),
		))

		first := m.values[0]
		last := m.values[len(m.values)-1]
		delta := last - first
		direction := "+"
		if delta < 0 {
			direction = ""
		}
		fmt.Printf("  %s: "+m.format+" -> "+m.format+" (%s"+m.format+")\n\n",
			m.name, first, last, direction, delta)
	}
}

func init() {
	progressCmd.Flags().StringVar(&progressRiotID, "riot-id", "", "Riot ID (format: gameName#tagLine, e.g., Faker#KR1)")
	progressCmd.Flags().BoolVar(&progressGraph, "graph", false, "Render ASCII graphs instead of JSON")
	rootCmd.AddCommand(progressCmd)
}
