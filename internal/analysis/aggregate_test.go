package analysis

import (
	"testing"

	"github.com/HatiCode/league-buddy/internal/models"
)

func makeAnalysisMatch(matchID, puuid, champion, role string, win bool, kills, deaths, assists int) models.Match {
	return models.Match{
		Metadata: models.MatchMetadata{MatchID: matchID},
		Info: models.MatchInfo{
			GameDuration: 1800,
			Participants: []models.Participant{
				{
					PUUID:                       puuid,
					ChampionName:                champion,
					TeamPosition:                role,
					TeamID:                      100,
					Kills:                       kills,
					Deaths:                      deaths,
					Assists:                     assists,
					TotalDamageDealtToChampions: 20000,
					TotalDamageTaken:            15000,
					TotalMinionsKilled:          200,
					NeutralMinionsKilled:        30,
					VisionScore:                 25,
					WardsPlaced:                 10,
					DetectorWardsPlaced:         2,
					TimeCCingOthers:             12,
					GoldEarned:                  13000,
					DamageDealtToBuildings:      2500,
					TotalTimeSpentDead:          60,
					Win:                         win,
				},
				{
					PUUID:                       "teammate",
					TeamID:                      100,
					Kills:                       3,
					TotalDamageDealtToChampions: 10000,
					TotalDamageTaken:            12000,
					DamageDealtToBuildings:      1500,
				},
				{
					PUUID:        "opponent",
					TeamID:       200,
					TeamPosition: role,
				},
			},
			Teams: []models.Team{
				{
					TeamID: 100,
					Objectives: models.TeamObjectives{
						Dragon:     models.ObjectiveStats{Kills: 2},
						Baron:      models.ObjectiveStats{Kills: 1},
						RiftHerald: models.ObjectiveStats{Kills: 1},
					},
				},
				{TeamID: 200},
			},
		},
	}
}

func TestAnalyzePlayerBasic(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 10, 3, 8),
		makeAnalysisMatch("M2", puuid, "Ahri", "MIDDLE", false, 2, 7, 3),
		makeAnalysisMatch("M3", puuid, "Zed", "MIDDLE", true, 8, 2, 5),
		makeAnalysisMatch("M4", puuid, "Lux", "BOTTOM", true, 5, 4, 12),
		makeAnalysisMatch("M5", puuid, "Ahri", "MIDDLE", true, 7, 1, 9),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:    puuid,
		GameName: "TestPlayer",
		TagLine:  "EUW",
		Matches:  matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PUUID != puuid {
		t.Errorf("puuid = %q, want %q", result.PUUID, puuid)
	}
	if result.TotalMatches != 5 {
		t.Errorf("totalMatches = %d, want 5", result.TotalMatches)
	}

	// Win rate: 4/5 = 0.8
	if !approxEqual(result.WinRate, 0.8) {
		t.Errorf("winRate = %f, want 0.8", result.WinRate)
	}
}

func TestAnalyzePlayerWinRate(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
		makeAnalysisMatch("M2", puuid, "Ahri", "MIDDLE", false, 3, 5, 2),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !approxEqual(result.WinRate, 0.5) {
		t.Errorf("winRate = %f, want 0.5", result.WinRate)
	}
}

func TestAnalyzePlayerChampionPool(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 8, 2, 6),
		makeAnalysisMatch("M2", puuid, "Ahri", "MIDDLE", true, 10, 1, 7),
		makeAnalysisMatch("M3", puuid, "Ahri", "MIDDLE", false, 3, 5, 4),
		makeAnalysisMatch("M4", puuid, "Zed", "MIDDLE", true, 12, 3, 2),
		makeAnalysisMatch("M5", puuid, "Lux", "BOTTOM", false, 2, 6, 8),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.ChampionPool) != 3 {
		t.Fatalf("championPool size = %d, want 3", len(result.ChampionPool))
	}

	// Sorted by games played: Ahri (3) > Zed (1) = Lux (1)
	if result.ChampionPool[0].ChampionName != "Ahri" {
		t.Errorf("top champion = %q, want Ahri", result.ChampionPool[0].ChampionName)
	}
	if result.ChampionPool[0].GamesPlayed != 3 {
		t.Errorf("Ahri gamesPlayed = %d, want 3", result.ChampionPool[0].GamesPlayed)
	}
	// Ahri win rate: 2/3
	if !approxEqual(result.ChampionPool[0].WinRate, 0.667) {
		t.Errorf("Ahri winRate = %f, want ~0.667", result.ChampionPool[0].WinRate)
	}
}

func TestAnalyzePlayerRoleBreakdown(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
		makeAnalysisMatch("M2", puuid, "Ahri", "MIDDLE", true, 7, 2, 6),
		makeAnalysisMatch("M3", puuid, "Lux", "BOTTOM", false, 2, 5, 8),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.RoleBreakdown) != 2 {
		t.Fatalf("roleBreakdown size = %d, want 2", len(result.RoleBreakdown))
	}

	// MIDDLE should be first (2 games > 1 game)
	if result.RoleBreakdown[0].Role != "MIDDLE" {
		t.Errorf("top role = %q, want MIDDLE", result.RoleBreakdown[0].Role)
	}
	if result.RoleBreakdown[0].GamesPlayed != 2 {
		t.Errorf("MIDDLE games = %d, want 2", result.RoleBreakdown[0].GamesPlayed)
	}
}

func TestAnalyzePlayerConsistency(t *testing.T) {
	puuid := "test-puuid"

	// All identical stats -> stddev should be 0
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
		makeAnalysisMatch("M2", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
		makeAnalysisMatch("M3", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Consistency.KDAStdDev != 0 {
		t.Errorf("KDAStdDev = %f, want 0 for identical matches", result.Consistency.KDAStdDev)
	}
	if result.Consistency.CSPerMinStdDev != 0 {
		t.Errorf("CSPerMinStdDev = %f, want 0 for identical matches", result.Consistency.CSPerMinStdDev)
	}
	if result.Consistency.DPMStdDev != 0 {
		t.Errorf("DPMStdDev = %f, want 0 for identical matches", result.Consistency.DPMStdDev)
	}
}

func TestAnalyzePlayerSingleMatch(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 10, 2, 8),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalMatches != 1 {
		t.Errorf("totalMatches = %d, want 1", result.TotalMatches)
	}
	if result.Consistency.KDAStdDev != 0 {
		t.Errorf("KDAStdDev = %f, want 0 for single match", result.Consistency.KDAStdDev)
	}
}

func TestAnalyzePlayerWithLeague(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
		League: &models.LeagueEntry{
			Tier:         "GOLD",
			Rank:         "II",
			LeaguePoints: 75,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Tier != "GOLD" {
		t.Errorf("tier = %q, want GOLD", result.Tier)
	}
	if result.Rank != "II" {
		t.Errorf("rank = %q, want II", result.Rank)
	}
	if result.LeaguePoints != 75 {
		t.Errorf("leaguePoints = %d, want 75", result.LeaguePoints)
	}
}

func TestAnalyzePlayerWithoutLeague(t *testing.T) {
	puuid := "test-puuid"
	matches := []models.Match{
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Tier != "" {
		t.Errorf("tier = %q, want empty", result.Tier)
	}
}

func TestAnalyzePlayerNoMatches(t *testing.T) {
	_, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID: "test",
	})
	if err == nil {
		t.Fatal("expected error for empty matches")
	}
}

func TestAnalyzePlayerSkipsRemakes(t *testing.T) {
	puuid := "test-puuid"
	remake := makeAnalysisMatch("REMAKE", puuid, "Ahri", "MIDDLE", false, 0, 0, 0)
	remake.Info.GameDuration = 30

	matches := []models.Match{
		remake,
		makeAnalysisMatch("M1", puuid, "Ahri", "MIDDLE", true, 5, 3, 5),
	}

	result, err := AnalyzePlayer(PlayerAnalysisParams{
		PUUID:   puuid,
		Matches: matches,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalMatches != 1 {
		t.Errorf("totalMatches = %d, want 1 (remake should be skipped)", result.TotalMatches)
	}
}

func TestIdentifyInsightsStrengths(t *testing.T) {
	avg := AverageMetrics{
		KDA:                    4.5,
		KillParticipation:      0.70,
		CSPerMinute:            8.0,
		VisionScorePerMinute:   1.5,
		DamageShare:            0.30,
		ObjectiveParticipation: 0.65,
		DeathsPerMinute:        0.10,
	}

	strengths, weaknesses := identifyInsights(avg, make([]ChampionStats, 5), ConsistencyMetrics{KDAStdDev: 0.5})

	if len(strengths) == 0 {
		t.Error("expected at least one strength")
	}
	if len(weaknesses) != 0 {
		t.Errorf("expected no weaknesses, got %d", len(weaknesses))
	}

	hasCategory := func(insights []Insight, cat string) bool {
		for _, i := range insights {
			if i.Category == cat {
				return true
			}
		}
		return false
	}

	if !hasCategory(strengths, "combat") {
		t.Error("expected combat strength")
	}
	if !hasCategory(strengths, "farming") {
		t.Error("expected farming strength")
	}
	if !hasCategory(strengths, "vision") {
		t.Error("expected vision strength")
	}
}

func TestIdentifyInsightsWeaknesses(t *testing.T) {
	avg := AverageMetrics{
		KDA:                    1.2,
		KillParticipation:      0.35,
		CSPerMinute:            4.5,
		VisionScorePerMinute:   0.4,
		DamageShare:            0.12,
		ObjectiveParticipation: 0.25,
		DeathsPerMinute:        0.30,
	}

	strengths, weaknesses := identifyInsights(avg, make([]ChampionStats, 1), ConsistencyMetrics{KDAStdDev: 4.0})

	if len(weaknesses) == 0 {
		t.Error("expected at least one weakness")
	}
	if len(strengths) != 0 {
		t.Errorf("expected no strengths, got %d", len(strengths))
	}
}

func TestStddev(t *testing.T) {
	// Known values: [2, 4, 4, 4, 5, 5, 7, 9] -> mean=5, stddev=2.0
	values := []float64{2, 4, 4, 4, 5, 5, 7, 9}
	result := stddev(values)
	if !approxEqual(result, 2.0) {
		t.Errorf("stddev = %f, want ~2.0", result)
	}
}

func TestStddevSingleValue(t *testing.T) {
	result := stddev([]float64{5.0})
	if result != 0 {
		t.Errorf("stddev of single value = %f, want 0", result)
	}
}
