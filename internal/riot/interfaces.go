package riot

import (
	"context"

	"github.com/HatiCode/league-buddy/internal/models"
)

// AccountFetcher retrieves Riot account information.
type AccountFetcher interface {
	GetAccountByRiotID(ctx context.Context, region, gameName, tagLine string) (*models.Account, error)
}

// SummonerFetcher retrieves summoner information.
type SummonerFetcher interface {
	GetSummonerByPUUID(ctx context.Context, platform, puuid string) (*models.Summoner, error)
}

// MatchFetcher retrieves match data.
type MatchFetcher interface {
	GetMatchIDs(ctx context.Context, region, puuid string, count int) ([]string, error)
	GetMatch(ctx context.Context, region, matchID string) (*models.Match, error)
	GetMatchTimeline(ctx context.Context, region, matchID string) (*models.Timeline, error)
}

// LeagueFetcher retrieves ranked/league information.
type LeagueFetcher interface {
	GetLeagueEntries(ctx context.Context, region, puuid string) ([]models.LeagueEntry, error)
}

// Client combines all Riot API capabilities.
// Consumers should prefer the smaller interfaces when possible.
type Client interface {
	AccountFetcher
	SummonerFetcher
	MatchFetcher
	LeagueFetcher
}
