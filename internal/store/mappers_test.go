package store_test

import (
	"testing"

	"github.com/HatiCode/league-buddy/internal/models"
	"github.com/HatiCode/league-buddy/internal/store"
)

func TestSummonerFromAPI(t *testing.T) {
	apiSummoner := &models.Summoner{
		ID:            "encrypted-id",
		AccountID:     "account-id",
		PUUID:         "puuid-12345",
		Name:          "TestPlayer",
		ProfileIconID: 1234,
		SummonerLevel: 100,
	}

	result := store.SummonerFromAPI(apiSummoner, "euw1")

	if result.PUUID != apiSummoner.PUUID {
		t.Errorf("expected PUUID %s, got %s", apiSummoner.PUUID, result.PUUID)
	}
	if result.SummonerID != apiSummoner.ID {
		t.Errorf("expected SummonerID %s, got %s", apiSummoner.ID, result.SummonerID)
	}
	if result.Name != apiSummoner.Name {
		t.Errorf("expected Name %s, got %s", apiSummoner.Name, result.Name)
	}
	if result.Platform != "euw1" {
		t.Errorf("expected Platform euw1, got %s", result.Platform)
	}
	if result.ProfileIconID != apiSummoner.ProfileIconID {
		t.Errorf("expected ProfileIconID %d, got %d", apiSummoner.ProfileIconID, result.ProfileIconID)
	}
	if result.SummonerLevel != apiSummoner.SummonerLevel {
		t.Errorf("expected SummonerLevel %d, got %d", apiSummoner.SummonerLevel, result.SummonerLevel)
	}
}

func TestSummoner_ApplyLeagueEntry(t *testing.T) {
	summoner := &store.Summoner{
		PUUID: "puuid-12345",
		Name:  "TestPlayer",
	}

	entry := &models.LeagueEntry{
		Tier:         "GOLD",
		Rank:         "II",
		LeaguePoints: 75,
	}

	summoner.ApplyLeagueEntry(entry)

	if summoner.Tier != "GOLD" {
		t.Errorf("expected Tier GOLD, got %s", summoner.Tier)
	}
	if summoner.Rank != "II" {
		t.Errorf("expected Rank II, got %s", summoner.Rank)
	}
	if summoner.LeaguePoints != 75 {
		t.Errorf("expected LeaguePoints 75, got %d", summoner.LeaguePoints)
	}
}

func TestMatchFromAPI(t *testing.T) {
	apiMatch := &models.Match{
		Metadata: models.MatchMetadata{
			MatchID:     "EUW1_12345",
			DataVersion: "2",
		},
		Info: models.MatchInfo{
			GameCreation: 1700000000000, // Unix ms
			GameDuration: 1800,          // 30 minutes in seconds
			GameMode:     "CLASSIC",
			QueueID:      420,
			PlatformID:   "EUW1",
			GameVersion:  "13.24.1",
		},
	}

	result := store.MatchFromAPI(apiMatch)

	if result.MatchID != "EUW1_12345" {
		t.Errorf("expected MatchID EUW1_12345, got %s", result.MatchID)
	}
	if result.Platform != "EUW1" {
		t.Errorf("expected Platform EUW1, got %s", result.Platform)
	}
	if result.QueueID != 420 {
		t.Errorf("expected QueueID 420, got %d", result.QueueID)
	}
	if result.GameMode != "CLASSIC" {
		t.Errorf("expected GameMode CLASSIC, got %s", result.GameMode)
	}
	if result.GameDuration != 1800 {
		t.Errorf("expected GameDuration 1800, got %d", result.GameDuration)
	}
	if result.GameVersion != "13.24.1" {
		t.Errorf("expected GameVersion 13.24.1, got %s", result.GameVersion)
	}
}

func TestParticipantsFromAPI(t *testing.T) {
	apiMatch := &models.Match{
		Info: models.MatchInfo{
			Participants: []models.Participant{
				{
					PUUID:                      "puuid-1",
					SummonerName:               "Player1",
					ChampionID:                 103,
					ChampionName:               "Ahri",
					TeamID:                     100,
					TeamPosition:               "MIDDLE",
					Win:                        true,
					Kills:                      10,
					Deaths:                     2,
					Assists:                    8,
					TotalMinionsKilled:         180,
					NeutralMinionsKilled:       20,
					VisionScore:                35,
					WardsPlaced:                12,
					WardsKilled:                5,
					DetectorWardsPlaced:        3,
					TotalDamageDealtToChampions: 25000,
					TotalDamageTaken:           15000,
					GoldEarned:                 14000,
					DragonKills:                1,
					BaronKills:                 0,
					TurretKills:                2,
					FirstBloodKill:             true,
					FirstBloodAssist:           false,
				},
				{
					PUUID:        "puuid-2",
					SummonerName: "Player2",
					ChampionName: "Jinx",
					TeamID:       100,
					TeamPosition: "BOTTOM",
					Win:          true,
				},
			},
		},
	}

	result := store.ParticipantsFromAPI(apiMatch)

	if len(result) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(result))
	}

	p1 := result[0]
	if p1.PUUID != "puuid-1" {
		t.Errorf("expected PUUID puuid-1, got %s", p1.PUUID)
	}
	if p1.ChampionName != "Ahri" {
		t.Errorf("expected ChampionName Ahri, got %s", p1.ChampionName)
	}
	if p1.Kills != 10 {
		t.Errorf("expected Kills 10, got %d", p1.Kills)
	}
	if p1.DamageDealt != 25000 {
		t.Errorf("expected DamageDealt 25000, got %d", p1.DamageDealt)
	}
	if !p1.FirstBloodKill {
		t.Error("expected FirstBloodKill to be true")
	}
}
