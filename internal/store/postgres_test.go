package store_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/HatiCode/league-buddy/internal/store"
)

// skipIfNoDatabase skips the test if DATABASE_URL is not set.
func skipIfNoDatabase(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	return dsn
}

func TestPostgres_UpsertAndGetSummoner(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	summoner := &store.Summoner{
		PUUID:         "test-puuid-" + time.Now().Format("20060102150405"),
		SummonerID:    "test-summoner-id",
		Name:          "TestPlayer",
		Platform:      "euw1",
		ProfileIconID: 1234,
		SummonerLevel: 100,
		Tier:          "GOLD",
		Rank:          "II",
		LeaguePoints:  50,
	}

	// Insert
	err = db.UpsertSummoner(ctx, summoner)
	if err != nil {
		t.Fatalf("UpsertSummoner failed: %v", err)
	}

	// Retrieve by PUUID
	retrieved, err := db.GetSummonerByPUUID(ctx, summoner.PUUID)
	if err != nil {
		t.Fatalf("GetSummonerByPUUID failed: %v", err)
	}
	if retrieved.Name != summoner.Name {
		t.Errorf("expected Name %s, got %s", summoner.Name, retrieved.Name)
	}
	if retrieved.ID == 0 {
		t.Error("expected ID to be set")
	}

	// Update
	summoner.Tier = "PLATINUM"
	summoner.LeaguePoints = 10
	err = db.UpsertSummoner(ctx, summoner)
	if err != nil {
		t.Fatalf("UpsertSummoner (update) failed: %v", err)
	}

	// Verify update
	retrieved, err = db.GetSummonerByPUUID(ctx, summoner.PUUID)
	if err != nil {
		t.Fatalf("GetSummonerByPUUID after update failed: %v", err)
	}
	if retrieved.Tier != "PLATINUM" {
		t.Errorf("expected Tier PLATINUM, got %s", retrieved.Tier)
	}
}

func TestPostgres_GetSummonerByName(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	summoner := &store.Summoner{
		PUUID:    "test-puuid-byname-" + time.Now().Format("20060102150405"),
		Name:     "UniqueTestName" + time.Now().Format("150405"),
		Platform: "euw1",
	}

	err = db.UpsertSummoner(ctx, summoner)
	if err != nil {
		t.Fatalf("UpsertSummoner failed: %v", err)
	}

	retrieved, err := db.GetSummonerByName(ctx, "euw1", summoner.Name)
	if err != nil {
		t.Fatalf("GetSummonerByName failed: %v", err)
	}
	if retrieved.PUUID != summoner.PUUID {
		t.Errorf("expected PUUID %s, got %s", summoner.PUUID, retrieved.PUUID)
	}
}

func TestPostgres_GetSummoner_NotFound(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	_, err = db.GetSummonerByPUUID(ctx, "non-existent-puuid")
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPostgres_SaveMatchAndParticipants(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	match := &store.Match{
		MatchID:      "TEST_" + time.Now().Format("20060102150405"),
		Platform:     "EUW1",
		QueueID:      420,
		GameMode:     "CLASSIC",
		GameDuration: 1800,
		GameVersion:  "13.24.1",
		GameEndedAt:  time.Now(),
	}

	participants := []store.Participant{
		{
			PUUID:        "part-puuid-1",
			SummonerName: "Player1",
			ChampionName: "Ahri",
			TeamID:       100,
			TeamPosition: "MIDDLE",
			Win:          true,
			Kills:        10,
			Deaths:       2,
			Assists:      8,
		},
		{
			PUUID:        "part-puuid-2",
			SummonerName: "Player2",
			ChampionName: "Jinx",
			TeamID:       100,
			TeamPosition: "BOTTOM",
			Win:          true,
			Kills:        8,
			Deaths:       3,
			Assists:      12,
		},
	}

	err = db.SaveMatch(ctx, match, participants)
	if err != nil {
		t.Fatalf("SaveMatch failed: %v", err)
	}

	// Verify match was saved
	retrieved, err := db.GetMatchByRiotID(ctx, match.MatchID)
	if err != nil {
		t.Fatalf("GetMatchByRiotID failed: %v", err)
	}
	if retrieved.GameDuration != 1800 {
		t.Errorf("expected GameDuration 1800, got %d", retrieved.GameDuration)
	}

	// Verify participants
	parts, err := db.GetParticipants(ctx, retrieved.ID)
	if err != nil {
		t.Fatalf("GetParticipants failed: %v", err)
	}
	if len(parts) != 2 {
		t.Errorf("expected 2 participants, got %d", len(parts))
	}
}

func TestPostgres_LinkSummonerMatch(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Create summoner
	summoner := &store.Summoner{
		PUUID:    "link-test-puuid-" + time.Now().Format("20060102150405"),
		Name:     "LinkTestPlayer",
		Platform: "euw1",
	}
	err = db.UpsertSummoner(ctx, summoner)
	if err != nil {
		t.Fatalf("UpsertSummoner failed: %v", err)
	}
	summoner, _ = db.GetSummonerByPUUID(ctx, summoner.PUUID)

	// Create match
	match := &store.Match{
		MatchID:      "LINK_TEST_" + time.Now().Format("20060102150405"),
		Platform:     "EUW1",
		QueueID:      420,
		GameMode:     "CLASSIC",
		GameDuration: 1800,
		GameEndedAt:  time.Now(),
	}
	err = db.SaveMatch(ctx, match, nil)
	if err != nil {
		t.Fatalf("SaveMatch failed: %v", err)
	}
	match, _ = db.GetMatchByRiotID(ctx, match.MatchID)

	// Link them
	err = db.LinkSummonerMatch(ctx, summoner.ID, match.ID)
	if err != nil {
		t.Fatalf("LinkSummonerMatch failed: %v", err)
	}

	// Verify link
	matches, err := db.GetMatchesForSummoner(ctx, summoner.ID)
	if err != nil {
		t.Fatalf("GetMatchesForSummoner failed: %v", err)
	}
	if len(matches) != 1 {
		t.Errorf("expected 1 match, got %d", len(matches))
	}
}

func TestPostgres_UnlinkOldestMatches(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Create summoner
	ts := time.Now().Format("20060102150405")
	summoner := &store.Summoner{
		PUUID:    "unlink-test-puuid-" + ts,
		Name:     "UnlinkTestPlayer",
		Platform: "euw1",
	}
	err = db.UpsertSummoner(ctx, summoner)
	if err != nil {
		t.Fatalf("UpsertSummoner failed: %v", err)
	}
	summoner, _ = db.GetSummonerByPUUID(ctx, summoner.PUUID)

	// Create and link 5 matches
	for i := 0; i < 5; i++ {
		match := &store.Match{
			MatchID:      "UNLINK_" + ts + "_" + string(rune('A'+i)),
			Platform:     "EUW1",
			QueueID:      420,
			GameMode:     "CLASSIC",
			GameDuration: 1800,
			GameEndedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		err = db.SaveMatch(ctx, match, nil)
		if err != nil {
			t.Fatalf("SaveMatch failed: %v", err)
		}
		match, _ = db.GetMatchByRiotID(ctx, match.MatchID)
		err = db.LinkSummonerMatch(ctx, summoner.ID, match.ID)
		if err != nil {
			t.Fatalf("LinkSummonerMatch failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different created_at
	}

	// Keep only 3
	err = db.UnlinkOldestMatches(ctx, summoner.ID, 3)
	if err != nil {
		t.Fatalf("UnlinkOldestMatches failed: %v", err)
	}

	// Verify only 3 remain
	matches, err := db.GetMatchesForSummoner(ctx, summoner.ID)
	if err != nil {
		t.Fatalf("GetMatchesForSummoner failed: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
}

func TestPostgres_DeleteOrphanedMatches(t *testing.T) {
	dsn := skipIfNoDatabase(t)
	ctx := context.Background()

	db, err := store.NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Create an orphaned match (no links)
	ts := time.Now().Format("20060102150405")
	match := &store.Match{
		MatchID:      "ORPHAN_" + ts,
		Platform:     "EUW1",
		QueueID:      420,
		GameMode:     "CLASSIC",
		GameDuration: 1800,
		GameEndedAt:  time.Now(),
	}
	err = db.SaveMatch(ctx, match, nil)
	if err != nil {
		t.Fatalf("SaveMatch failed: %v", err)
	}

	// Run cleanup
	deleted, err := db.DeleteOrphanedMatches(ctx)
	if err != nil {
		t.Fatalf("DeleteOrphanedMatches failed: %v", err)
	}
	if deleted < 1 {
		t.Errorf("expected at least 1 deleted, got %d", deleted)
	}

	// Verify it's gone
	_, err = db.GetMatchByRiotID(ctx, match.MatchID)
	if err != store.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
