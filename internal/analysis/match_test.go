package analysis

import (
	"math"
	"testing"

	"github.com/HatiCode/league-buddy/internal/models"
)

const testPUUID = "test-puuid-123"

func makeMatch(matchID, puuid string, opts ...func(*models.Participant)) *models.Match {
	participant := models.Participant{
		PUUID:                       puuid,
		ChampionName:                "Ahri",
		TeamPosition:                "MIDDLE",
		TeamID:                      100,
		Kills:                       10,
		Deaths:                      3,
		Assists:                     8,
		TotalDamageDealtToChampions: 25000,
		TotalDamageTaken:            18000,
		TotalMinionsKilled:          180,
		NeutralMinionsKilled:        20,
		VisionScore:                 30,
		WardsPlaced:                 12,
		DetectorWardsPlaced:         3,
		TimeCCingOthers:             15,
		GoldEarned:                  14000,
		DamageDealtToBuildings:      3000,
		TotalTimeSpentDead:          90,
		Win:                         true,
	}
	for _, opt := range opts {
		opt(&participant)
	}

	teammate := models.Participant{
		PUUID:                       "teammate-1",
		TeamID:                      100,
		Kills:                       5,
		Deaths:                      4,
		Assists:                     10,
		TotalDamageDealtToChampions: 15000,
		TotalDamageTaken:            22000,
		DamageDealtToBuildings:      2000,
	}

	opponent := models.Participant{
		PUUID:                       "opponent-1",
		TeamID:                      200,
		TeamPosition:                "MIDDLE",
		Kills:                       4,
		Deaths:                      6,
		Assists:                     5,
		TotalDamageDealtToChampions: 12000,
		TotalDamageTaken:            20000,
		DamageDealtToBuildings:      1000,
	}

	return &models.Match{
		Metadata: models.MatchMetadata{MatchID: matchID},
		Info: models.MatchInfo{
			GameDuration: 1800, // 30 minutes
			Participants: []models.Participant{participant, teammate, opponent},
			Teams: []models.Team{
				{
					TeamID: 100,
					Objectives: models.TeamObjectives{
						Dragon:     models.ObjectiveStats{Kills: 3},
						Baron:      models.ObjectiveStats{Kills: 1},
						RiftHerald: models.ObjectiveStats{Kills: 1},
					},
				},
				{TeamID: 200},
			},
		},
	}
}

func TestAnalyzeMatchWithChallenges(t *testing.T) {
	match := makeMatch("EUW1_123", testPUUID, func(p *models.Participant) {
		p.Challenges = &models.Challenges{
			KDA:                              6.0,
			KillParticipation:                0.72,
			DamagePerMinute:                  833.3,
			TeamDamagePercentage:             0.35,
			GoldPerMinute:                    466.7,
			VisionScorePerMinute:             1.0,
			DamageTakenOnTeamPercentage:      0.20,
			EffectiveHealAndShielding:        1500,
			SoloKills:                        3,
			ControlWardsPlaced:               4,
			LaneMinionsFirst10Minutes:        80,
			EarlyLaningPhaseGoldExpAdvantage: 500,
			LaningPhaseGoldExpAdvantage:      800,
			MaxCsAdvantageOnLaneOpponent:     15.0,
			DragonTakedowns:                  2,
			BaronTakedowns:                   1,
			RiftHeraldTakedowns:              1,
		}
	})

	result, err := AnalyzeMatch(match, testPUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.Metrics
	if m.MatchID != "EUW1_123" {
		t.Errorf("matchID = %q, want %q", m.MatchID, "EUW1_123")
	}
	if m.ChampionName != "Ahri" {
		t.Errorf("championName = %q, want %q", m.ChampionName, "Ahri")
	}
	if m.Role != "MIDDLE" {
		t.Errorf("role = %q, want %q", m.Role, "MIDDLE")
	}
	if !m.Win {
		t.Error("win = false, want true")
	}

	if m.KDA != 6.0 {
		t.Errorf("KDA = %f, want 6.0", m.KDA)
	}
	if m.KillParticipation != 0.72 {
		t.Errorf("killParticipation = %f, want 0.72", m.KillParticipation)
	}
	if m.DamagePerMinute != 833.3 {
		t.Errorf("DPM = %f, want 833.3", m.DamagePerMinute)
	}
	if m.DamageShare != 0.35 {
		t.Errorf("damageShare = %f, want 0.35", m.DamageShare)
	}
	if m.SoloKills != 3 {
		t.Errorf("soloKills = %d, want 3", m.SoloKills)
	}

	// Objective participation: (2+1+1) / (3+1+1) = 4/5 = 0.8
	if !approxEqual(m.ObjectiveParticipation, 0.8) {
		t.Errorf("objectiveParticipation = %f, want ~0.8", m.ObjectiveParticipation)
	}

	// CS/min: (180+20) / 30 = 6.67
	if !approxEqual(m.CSPerMinute, 6.667) {
		t.Errorf("CSPerMinute = %f, want ~6.667", m.CSPerMinute)
	}

	// Wards/min: 12 / 30 = 0.4
	if !approxEqual(m.WardsPerMinute, 0.4) {
		t.Errorf("wardsPerMinute = %f, want ~0.4", m.WardsPerMinute)
	}

	// Deaths/min: 3 / 30 = 0.1
	if !approxEqual(m.DeathsPerMinute, 0.1) {
		t.Errorf("deathsPerMinute = %f, want ~0.1", m.DeathsPerMinute)
	}

	// Turret damage share: 3000 / (3000+2000) = 0.6
	if !approxEqual(m.TurretDamageShare, 0.6) {
		t.Errorf("turretDamageShare = %f, want ~0.6", m.TurretDamageShare)
	}

	if result.LanePhase != nil {
		t.Error("lanePhase should be nil when no timeline provided")
	}
}

func TestAnalyzeMatchWithoutChallenges(t *testing.T) {
	match := makeMatch("EUW1_456", testPUUID)

	result, err := AnalyzeMatch(match, testPUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m := result.Metrics

	// KDA fallback: (10+8) / 3 = 6.0
	if !approxEqual(m.KDA, 6.0) {
		t.Errorf("KDA = %f, want ~6.0", m.KDA)
	}

	// Kill participation: (10+8) / (10+5) = 18/15 = 1.2 (can exceed 1.0 due to assist overlap)
	if !approxEqual(m.KillParticipation, 1.2) {
		t.Errorf("killParticipation = %f, want ~1.2", m.KillParticipation)
	}

	// DPM: 25000 / 30 = 833.33
	if !approxEqual(m.DamagePerMinute, 833.333) {
		t.Errorf("DPM = %f, want ~833.33", m.DamagePerMinute)
	}

	// Damage share: 25000 / (25000+15000) = 0.625
	if !approxEqual(m.DamageShare, 0.625) {
		t.Errorf("damageShare = %f, want ~0.625", m.DamageShare)
	}

	// Damage taken share: 18000 / (18000+22000) = 0.45
	if !approxEqual(m.DamageTakenShare, 0.45) {
		t.Errorf("damageTakenShare = %f, want ~0.45", m.DamageTakenShare)
	}
}

func TestAnalyzeMatchZeroDeaths(t *testing.T) {
	match := makeMatch("EUW1_789", testPUUID, func(p *models.Participant) {
		p.Deaths = 0
	})

	result, err := AnalyzeMatch(match, testPUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With 0 deaths, fallback uses 1 as denominator: (10+8)/1 = 18
	if !approxEqual(result.Metrics.KDA, 18.0) {
		t.Errorf("KDA with 0 deaths = %f, want 18.0", result.Metrics.KDA)
	}

	if result.Metrics.DeathsPerMinute != 0 {
		t.Errorf("deathsPerMinute = %f, want 0", result.Metrics.DeathsPerMinute)
	}
}

func TestAnalyzeMatchRemake(t *testing.T) {
	match := makeMatch("EUW1_REMAKE", testPUUID)
	match.Info.GameDuration = 30

	_, err := AnalyzeMatch(match, testPUUID)
	if err != ErrMatchTooShort {
		t.Errorf("err = %v, want ErrMatchTooShort", err)
	}
}

func TestAnalyzeMatchParticipantNotFound(t *testing.T) {
	match := makeMatch("EUW1_NF", testPUUID)

	_, err := AnalyzeMatch(match, "nonexistent-puuid")
	if err == nil {
		t.Fatal("expected error for missing participant")
	}
}

func TestAnalyzeMatchZeroTeamObjectives(t *testing.T) {
	match := makeMatch("EUW1_NOOBJ", testPUUID, func(p *models.Participant) {
		p.Challenges = &models.Challenges{
			DragonTakedowns:     0,
			BaronTakedowns:      0,
			RiftHeraldTakedowns: 0,
		}
	})
	match.Info.Teams[0].Objectives = models.TeamObjectives{}

	result, err := AnalyzeMatch(match, testPUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metrics.ObjectiveParticipation != 0 {
		t.Errorf("objectiveParticipation = %f, want 0", result.Metrics.ObjectiveParticipation)
	}
}

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}
