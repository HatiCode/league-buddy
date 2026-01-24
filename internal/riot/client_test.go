package riot_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HatiCode/league-buddy/internal/models"
	"github.com/HatiCode/league-buddy/internal/riot"
	"github.com/HatiCode/league-buddy/pkg/ratelimit"
)

// --- Summoner Fetcher Tests ---

func TestGetSummonerByName_Success(t *testing.T) {
	expected := &models.Summoner{
		ID:            "encrypted-summoner-id",
		AccountID:     "encrypted-account-id",
		PUUID:         "puuid-12345",
		Name:          "Faker",
		ProfileIconID: 4567,
		SummonerLevel: 500,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/lol/summoner/v4/summoners/by-name/Faker" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Riot-Token") == "" {
			t.Error("missing API key header")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	summoner, err := client.GetSummonerByName(context.Background(), riot.PlatformEUW1, "Faker")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summoner.PUUID != expected.PUUID {
		t.Errorf("expected PUUID %s, got %s", expected.PUUID, summoner.PUUID)
	}
	if summoner.Name != expected.Name {
		t.Errorf("expected Name %s, got %s", expected.Name, summoner.Name)
	}
}

func TestGetSummonerByName_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":{"message":"Data not found","status_code":404}}`))
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	_, err := client.GetSummonerByName(context.Background(), riot.PlatformEUW1, "NonExistentPlayer12345")

	if err != riot.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGetSummonerByName_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"status":{"message":"Unauthorized","status_code":401}}`))
	}))
	defer server.Close()

	client := riot.NewClient("invalid-api-key", riot.WithBaseURL(server.URL))
	_, err := client.GetSummonerByName(context.Background(), riot.PlatformEUW1, "Faker")

	if err != riot.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGetSummonerByPUUID_Success(t *testing.T) {
	expected := &models.Summoner{
		ID:            "encrypted-summoner-id",
		AccountID:     "encrypted-account-id",
		PUUID:         "puuid-12345",
		Name:          "Faker",
		ProfileIconID: 4567,
		SummonerLevel: 500,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lol/summoner/v4/summoners/by-puuid/puuid-12345" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	summoner, err := client.GetSummonerByPUUID(context.Background(), riot.PlatformEUW1, "puuid-12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summoner.Name != expected.Name {
		t.Errorf("expected Name %s, got %s", expected.Name, summoner.Name)
	}
}

// --- Match Fetcher Tests ---

func TestGetMatchIDs_Success(t *testing.T) {
	expected := []string{"EUW1_12345", "EUW1_12346", "EUW1_12347", "EUW1_12348", "EUW1_12349"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lol/match/v5/matches/by-puuid/puuid-12345/ids" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		// Verify count query param
		if r.URL.Query().Get("count") != "5" {
			t.Errorf("expected count=5, got %s", r.URL.Query().Get("count"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	matchIDs, err := client.GetMatchIDs(context.Background(), riot.PlatformEUW1, "puuid-12345", 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matchIDs) != 5 {
		t.Errorf("expected 5 match IDs, got %d", len(matchIDs))
	}
	if matchIDs[0] != expected[0] {
		t.Errorf("expected first match ID %s, got %s", expected[0], matchIDs[0])
	}
}

func TestGetMatch_Success(t *testing.T) {
	expected := &models.Match{
		Metadata: models.MatchMetadata{
			MatchID:      "EUW1_12345",
			DataVersion:  "2",
			Participants: []string{"puuid-1", "puuid-2"},
		},
		Info: models.MatchInfo{
			GameDuration: 1800,
			GameMode:     "CLASSIC",
			QueueID:      420, // Ranked Solo
			Participants: []models.Participant{
				{
					PUUID:        "puuid-1",
					SummonerName: "Player1",
					ChampionName: "Ahri",
					Kills:        10,
					Deaths:       2,
					Assists:      8,
					Win:          true,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lol/match/v5/matches/EUW1_12345" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	match, err := client.GetMatch(context.Background(), riot.PlatformEUW1, "EUW1_12345")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match.Metadata.MatchID != expected.Metadata.MatchID {
		t.Errorf("expected match ID %s, got %s", expected.Metadata.MatchID, match.Metadata.MatchID)
	}
	if len(match.Info.Participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(match.Info.Participants))
	}
}

// --- League Fetcher Tests ---

func TestGetLeagueEntries_Success(t *testing.T) {
	expected := []models.LeagueEntry{
		{
			SummonerID:   "encrypted-summoner-id",
			SummonerName: "Faker",
			QueueType:    models.QueueRankedSolo,
			Tier:         "CHALLENGER",
			Rank:         "I",
			LeaguePoints: 1000,
			Wins:         200,
			Losses:       50,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/lol/league/v4/entries/by-summoner/encrypted-summoner-id" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	entries, err := client.GetLeagueEntries(context.Background(), riot.PlatformEUW1, "encrypted-summoner-id")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Tier != "CHALLENGER" {
		t.Errorf("expected tier CHALLENGER, got %s", entries[0].Tier)
	}
}

// --- Rate Limiting Tests ---

func TestClient_RateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "5")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"status":{"message":"Rate limit exceeded","status_code":429}}`))
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))
	_, err := client.GetSummonerByName(context.Background(), riot.PlatformEUW1, "Faker")

	if err != riot.ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

// --- Context Cancellation Tests ---

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		<-r.Context().Done()
	}))
	defer server.Close()

	client := riot.NewClient("test-api-key", riot.WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetSummonerByName(ctx, riot.PlatformEUW1, "Faker")

	if err == nil {
		t.Error("expected context cancellation error")
	}
}

// --- Region Validation Tests ---

func TestClient_InvalidRegion(t *testing.T) {
	client := riot.NewClient("test-api-key")
	_, err := client.GetSummonerByName(context.Background(), "invalid-region", "Faker")

	if err != riot.ErrInvalidRegion {
		t.Errorf("expected ErrInvalidRegion, got %v", err)
	}
}

// --- Rate Limiter Integration Tests ---

func TestClient_WithRateLimiter(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&models.Summoner{Name: "Test"})
	}))
	defer server.Close()

	limiter := ratelimit.NewLimiter(
		ratelimit.WithLimit(2, 100*time.Millisecond),
	)

	client := riot.NewClient("test-api-key",
		riot.WithBaseURL(server.URL),
		riot.WithRateLimiter(limiter),
	)

	// First two requests should succeed immediately
	for i := 0; i < 2; i++ {
		_, err := client.GetSummonerByName(context.Background(), riot.PlatformEUW1, "Test")
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
	}

	// Third request should block and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.GetSummonerByName(ctx, riot.PlatformEUW1, "Test")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests to server, got %d", requestCount)
	}
}
