package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore creates a new PostgreSQL store.
func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

// Close closes the database connection.
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// DB returns the underlying *sql.DB for migrations.
func (s *PostgresStore) DB() *sql.DB {
	return s.db.DB
}

// --- Summoner operations ---

func (s *PostgresStore) GetSummonerByPUUID(ctx context.Context, puuid string) (*Summoner, error) {
	var summoner Summoner
	err := s.db.GetContext(ctx, &summoner, `
		SELECT id, puuid, game_name, tag_line, platform, profile_icon_id, summoner_level,
		       revision_date, tier, rank, league_points, created_at, updated_at
		FROM summoners WHERE puuid = $1
	`, puuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &summoner, nil
}

func (s *PostgresStore) GetSummonerByRiotID(ctx context.Context, platform, gameName, tagLine string) (*Summoner, error) {
	var summoner Summoner
	err := s.db.GetContext(ctx, &summoner, `
		SELECT id, puuid, game_name, tag_line, platform, profile_icon_id, summoner_level,
		       revision_date, tier, rank, league_points, created_at, updated_at
		FROM summoners WHERE platform = $1 AND LOWER(game_name) = LOWER($2) AND LOWER(tag_line) = LOWER($3)
	`, platform, gameName, tagLine)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &summoner, nil
}

func (s *PostgresStore) UpsertSummoner(ctx context.Context, summoner *Summoner) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO summoners (puuid, game_name, tag_line, platform, profile_icon_id, summoner_level, revision_date, tier, rank, league_points, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (puuid) DO UPDATE SET
			game_name = EXCLUDED.game_name,
			tag_line = EXCLUDED.tag_line,
			profile_icon_id = EXCLUDED.profile_icon_id,
			summoner_level = EXCLUDED.summoner_level,
			revision_date = EXCLUDED.revision_date,
			tier = EXCLUDED.tier,
			rank = EXCLUDED.rank,
			league_points = EXCLUDED.league_points,
			updated_at = NOW()
	`, summoner.PUUID, summoner.GameName, summoner.TagLine, summoner.Platform,
		summoner.ProfileIconID, summoner.SummonerLevel, summoner.RevisionDate, summoner.Tier, summoner.Rank, summoner.LeaguePoints)
	return err
}

// --- Match operations ---

func (s *PostgresStore) GetMatchByRiotID(ctx context.Context, matchID string) (*Match, error) {
	var match Match
	err := s.db.GetContext(ctx, &match, `
		SELECT id, match_id, platform, queue_id, game_mode, game_duration, game_version, game_ended_at, created_at
		FROM matches WHERE match_id = $1
	`, matchID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (s *PostgresStore) GetMatchesForSummoner(ctx context.Context, summonerID int64) ([]Match, error) {
	var matches []Match
	err := s.db.SelectContext(ctx, &matches, `
		SELECT m.id, m.match_id, m.platform, m.queue_id, m.game_mode, m.game_duration, m.game_version, m.game_ended_at, m.created_at
		FROM matches m
		JOIN summoner_matches sm ON m.id = sm.match_id
		WHERE sm.summoner_id = $1
		ORDER BY sm.created_at DESC
	`, summonerID)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

func (s *PostgresStore) GetParticipants(ctx context.Context, matchID int64) ([]Participant, error) {
	var participants []Participant
	err := s.db.SelectContext(ctx, &participants, `
		SELECT id, match_id, puuid, summoner_name, champion_id, champion_name, team_id, team_position,
		       win, kills, deaths, assists, total_minions_killed, neutral_minions_killed,
		       vision_score, wards_placed, wards_killed, detector_wards_placed,
		       damage_dealt, damage_taken, gold_earned, dragon_kills, baron_kills, turret_kills,
		       first_blood_kill, first_blood_assist
		FROM participants WHERE match_id = $1
	`, matchID)
	if err != nil {
		return nil, err
	}
	return participants, nil
}

func (s *PostgresStore) GetParticipantByPUUID(ctx context.Context, matchID int64, puuid string) (*Participant, error) {
	var participant Participant
	err := s.db.GetContext(ctx, &participant, `
		SELECT id, match_id, puuid, summoner_name, champion_id, champion_name, team_id, team_position,
		       win, kills, deaths, assists, total_minions_killed, neutral_minions_killed,
		       vision_score, wards_placed, wards_killed, detector_wards_placed,
		       damage_dealt, damage_taken, gold_earned, dragon_kills, baron_kills, turret_kills,
		       first_blood_kill, first_blood_assist
		FROM participants WHERE match_id = $1 AND puuid = $2
	`, matchID, puuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &participant, nil
}

func (s *PostgresStore) SaveMatch(ctx context.Context, match *Match, participants []Participant) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Insert match
	err = tx.QueryRowxContext(ctx, `
		INSERT INTO matches (match_id, platform, queue_id, game_mode, game_duration, game_version, game_ended_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (match_id) DO UPDATE SET match_id = EXCLUDED.match_id
		RETURNING id
	`, match.MatchID, match.Platform, match.QueueID, match.GameMode, match.GameDuration, match.GameVersion, match.GameEndedAt).Scan(&match.ID)
	if err != nil {
		return err
	}

	// Insert participants
	for i := range participants {
		participants[i].MatchID = match.ID
		_, err = tx.ExecContext(ctx, `
			INSERT INTO participants (match_id, puuid, summoner_name, champion_id, champion_name, team_id, team_position,
			                          win, kills, deaths, assists, total_minions_killed, neutral_minions_killed,
			                          vision_score, wards_placed, wards_killed, detector_wards_placed,
			                          damage_dealt, damage_taken, gold_earned, dragon_kills, baron_kills, turret_kills,
			                          first_blood_kill, first_blood_assist)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
			ON CONFLICT (match_id, puuid) DO NOTHING
		`, participants[i].MatchID, participants[i].PUUID, participants[i].SummonerName,
			participants[i].ChampionID, participants[i].ChampionName, participants[i].TeamID, participants[i].TeamPosition,
			participants[i].Win, participants[i].Kills, participants[i].Deaths, participants[i].Assists,
			participants[i].TotalMinionsKilled, participants[i].NeutralMinionsKilled,
			participants[i].VisionScore, participants[i].WardsPlaced, participants[i].WardsKilled, participants[i].DetectorWardsPlaced,
			participants[i].DamageDealt, participants[i].DamageTaken, participants[i].GoldEarned,
			participants[i].DragonKills, participants[i].BaronKills, participants[i].TurretKills,
			participants[i].FirstBloodKill, participants[i].FirstBloodAssist)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) LinkSummonerMatch(ctx context.Context, summonerID, matchID int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO summoner_matches (summoner_id, match_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (summoner_id, match_id) DO NOTHING
	`, summonerID, matchID)
	return err
}

func (s *PostgresStore) UnlinkOldestMatches(ctx context.Context, summonerID int64, keepCount int) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM summoner_matches
		WHERE summoner_id = $1
		AND match_id NOT IN (
			SELECT match_id FROM summoner_matches
			WHERE summoner_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		)
	`, summonerID, keepCount)
	return err
}

// --- Coaching session operations ---

func (s *PostgresStore) GetLatestCoachingSession(ctx context.Context, puuid string) (*CoachingSession, error) {
	var session CoachingSession
	err := s.db.GetContext(ctx, &session, `
		SELECT id, puuid, latest_match_id, match_ids, analysis, advice, created_at
		FROM coaching_sessions
		WHERE puuid = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, puuid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *PostgresStore) GetCoachingSessions(ctx context.Context, puuid string) ([]CoachingSession, error) {
	var sessions []CoachingSession
	err := s.db.SelectContext(ctx, &sessions, `
		SELECT id, puuid, latest_match_id, match_ids, analysis, advice, created_at
		FROM coaching_sessions
		WHERE puuid = $1
		ORDER BY created_at ASC
	`, puuid)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *PostgresStore) SaveCoachingSession(ctx context.Context, session *CoachingSession) error {
	return s.db.QueryRowxContext(ctx, `
		INSERT INTO coaching_sessions (puuid, latest_match_id, match_ids, analysis, advice, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`, session.PUUID, session.LatestMatchID, session.MatchIDs, session.Analysis, session.Advice).
		Scan(&session.ID, &session.CreatedAt)
}

// --- Cleanup operations ---

func (s *PostgresStore) DeleteOrphanedMatches(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM matches
		WHERE id NOT IN (SELECT DISTINCT match_id FROM summoner_matches)
	`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
