package riot

import (
	"context"

	"github.com/HatiCode/league-buddy/internal/models"
)

// SummonerFetcher retrieves summoner information.
type SummonerFetcher interface {
	GetSummonerByName(ctx context.Context, region, name string) (*models.Summoner, error)
	GetSummonerByPUUID(ctx context.Context, region, puuid string) (*models.Summoner, error)
}

// MatchFetcher retrieves match data.
type MatchFetcher interface {
	GetMatchIDs(ctx context.Context, region, puuid string, count int) ([]string, error)
	GetMatch(ctx context.Context, region, matchID string) (*models.Match, error)
}

// LeagueFetcher retrieves ranked/league information.
type LeagueFetcher interface {
	GetLeagueEntries(ctx context.Context, region, summonerID string) ([]models.LeagueEntry, error)
}

// Client combines all Riot API capabilities.
// Consumers should prefer the smaller interfaces when possible.
type Client interface {
	SummonerFetcher
	MatchFetcher
	LeagueFetcher
}
