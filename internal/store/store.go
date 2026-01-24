package store

import "context"

// SummonerReader retrieves summoner data.
type SummonerReader interface {
	GetSummonerByPUUID(ctx context.Context, puuid string) (*Summoner, error)
	GetSummonerByName(ctx context.Context, platform, name string) (*Summoner, error)
}

// SummonerWriter persists summoner data.
type SummonerWriter interface {
	UpsertSummoner(ctx context.Context, summoner *Summoner) error
}

// SummonerRepository combines read and write operations for summoners.
type SummonerRepository interface {
	SummonerReader
	SummonerWriter
}

// MatchReader retrieves match data.
type MatchReader interface {
	GetMatchByRiotID(ctx context.Context, matchID string) (*Match, error)
	GetMatchesForSummoner(ctx context.Context, summonerID int64) ([]Match, error)
	GetParticipants(ctx context.Context, matchID int64) ([]Participant, error)
	GetParticipantByPUUID(ctx context.Context, matchID int64, puuid string) (*Participant, error)
}

// MatchWriter persists match data.
type MatchWriter interface {
	SaveMatch(ctx context.Context, match *Match, participants []Participant) error
	LinkSummonerMatch(ctx context.Context, summonerID, matchID int64) error
	UnlinkOldestMatches(ctx context.Context, summonerID int64, keepCount int) error
}

// MatchRepository combines read and write operations for matches.
type MatchRepository interface {
	MatchReader
	MatchWriter
}

// CleanupService handles orphaned data removal.
type CleanupService interface {
	DeleteOrphanedMatches(ctx context.Context) (int64, error)
}

// Store combines all repository interfaces.
type Store interface {
	SummonerRepository
	MatchRepository
	CleanupService
}
