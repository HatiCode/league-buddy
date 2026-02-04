package coaching

import (
	"strings"
	"testing"

	"github.com/HatiCode/league-buddy/internal/analysis"
)

func makeTestAnalysis() *analysis.PlayerAnalysis {
	return &analysis.PlayerAnalysis{
		PUUID:        "test-puuid",
		GameName:     "TestPlayer",
		TagLine:      "EUW",
		Tier:         "GOLD",
		Rank:         "II",
		LeaguePoints: 75,
		WinRate:      0.60,
		TotalMatches: 10,
		Averages: analysis.AverageMetrics{
			KDA:                    3.5,
			KillParticipation:      0.62,
			CSPerMinute:            6.8,
			DamagePerMinute:        750,
			DamageShare:            0.28,
			VisionScorePerMinute:   1.1,
			DeathsPerMinute:        0.15,
			GoldPerMinute:          420,
			ObjectiveParticipation: 0.55,
		},
		Consistency: analysis.ConsistencyMetrics{
			KDAStdDev:      1.2,
			CSPerMinStdDev: 0.8,
			DPMStdDev:      150,
		},
		ChampionPool: []analysis.ChampionStats{
			{ChampionName: "Ahri", GamesPlayed: 5, WinRate: 0.80, AvgKDA: 4.2},
			{ChampionName: "Zed", GamesPlayed: 3, WinRate: 0.33, AvgKDA: 2.5},
			{ChampionName: "Lux", GamesPlayed: 2, WinRate: 0.50, AvgKDA: 3.0},
		},
		RoleBreakdown: []analysis.RoleStats{
			{Role: "MIDDLE", GamesPlayed: 8, WinRate: 0.625},
			{Role: "BOTTOM", GamesPlayed: 2, WinRate: 0.50},
		},
		Strengths: []analysis.Insight{
			{Category: "combat", Description: "Strong KDA averaging 3.5", Value: 3.5, IsStrength: true},
		},
		Weaknesses: []analysis.Insight{
			{Category: "vision", Description: "Low vision score at 1.10 per minute", Value: 1.1},
		},
		Matches: []analysis.MatchAnalysis{
			{Metrics: analysis.MatchMetrics{
				MatchID: "EUW1_001", ChampionName: "Ahri", Role: "MIDDLE",
				KDA: 5.0, CSPerMinute: 7.2, DamagePerMinute: 800, Win: true,
			}},
			{Metrics: analysis.MatchMetrics{
				MatchID: "EUW1_002", ChampionName: "Zed", Role: "MIDDLE",
				KDA: 1.5, CSPerMinute: 6.0, DamagePerMinute: 650, Win: false,
			}},
		},
	}
}

func TestBuildInitialSystemPromptContainsSections(t *testing.T) {
	a := makeTestAnalysis()
	prompt := BuildInitialSystemPrompt(a)

	sections := []string{
		"League of Legends coach",
		"TestPlayer#EUW",
		"GOLD II",
		"60%",
		"KDA: 3.50",
		"CS/min: 6.8",
		"Ahri: 5 games",
		"MIDDLE: 8 games",
		"Strong KDA",
		"Low vision score",
		"Response Format",
		"Top 3 action items",
	}

	for _, section := range sections {
		if !strings.Contains(prompt, section) {
			t.Errorf("initial prompt missing %q", section)
		}
	}
}

func TestBuildInitialSystemPromptMatchHistory(t *testing.T) {
	a := makeTestAnalysis()
	prompt := BuildInitialSystemPrompt(a)

	if !strings.Contains(prompt, "Ahri MIDDLE (Win)") {
		t.Error("prompt missing Ahri win match")
	}
	if !strings.Contains(prompt, "Zed MIDDLE (Loss)") {
		t.Error("prompt missing Zed loss match")
	}
}

func TestBuildFollowUpSystemPromptContainsDelta(t *testing.T) {
	current := makeTestAnalysis()
	previous := makeTestAnalysis()
	previous.Averages.KDA = 2.5
	previous.Averages.CSPerMinute = 5.5
	previous.Averages.DeathsPerMinute = 0.25

	prompt := BuildFollowUpSystemPrompt(current, previous, "Focus on CS and reduce deaths.")

	sections := []string{
		"follow-up session",
		"Previous Session",
		"Progress Since Last Session",
		"KDA: 2.50 -> 3.50 (improved)",
		"CS/min: 5.5 -> 6.8 (improved)",
		"Deaths/min: 0.25 -> 0.15 (improved)",
		"Focus on CS and reduce deaths.",
		"Progress assessment",
		"Persistent weaknesses",
	}

	for _, section := range sections {
		if !strings.Contains(prompt, section) {
			t.Errorf("follow-up prompt missing %q", section)
		}
	}
}

func TestBuildFollowUpSystemPromptRegression(t *testing.T) {
	current := makeTestAnalysis()
	current.Averages.KDA = 2.0

	previous := makeTestAnalysis()
	previous.Averages.KDA = 3.5

	prompt := BuildFollowUpSystemPrompt(current, previous, "Previous advice.")

	if !strings.Contains(prompt, "KDA: 3.50 -> 2.00 (regressed)") {
		t.Error("follow-up prompt missing KDA regression")
	}
}

func TestBuildUserPrompt(t *testing.T) {
	initial := BuildUserPrompt(false)
	if !strings.Contains(initial, "Analyze") {
		t.Error("initial user prompt missing 'Analyze'")
	}

	followUp := BuildUserPrompt(true)
	if !strings.Contains(followUp, "follow-up") {
		t.Error("follow-up user prompt missing 'follow-up'")
	}
}

func TestBuildInitialSystemPromptNoRank(t *testing.T) {
	a := makeTestAnalysis()
	a.Tier = ""
	a.Rank = ""

	prompt := BuildInitialSystemPrompt(a)
	if strings.Contains(prompt, "Rank:") {
		t.Error("prompt should not contain rank line when tier is empty")
	}
}

func TestBuildInitialSystemPromptTokenBudget(t *testing.T) {
	a := makeTestAnalysis()
	prompt := BuildInitialSystemPrompt(a)

	// Rough estimate: 1 token ~= 4 chars. Prompt should stay under ~4K tokens (~16K chars)
	if len(prompt) > 16000 {
		t.Errorf("prompt is %d chars, likely exceeds 4K token budget", len(prompt))
	}
}
