package store

import "time"

// Summoner represents a stored player profile.
type Summoner struct {
	ID            int64     `db:"id"`
	PUUID         string    `db:"puuid"`
	GameName      string    `db:"game_name"`
	TagLine       string    `db:"tag_line"`
	Platform      string    `db:"platform"`
	ProfileIconID int       `db:"profile_icon_id"`
	SummonerLevel int64     `db:"summoner_level"`
	RevisionDate  int64     `db:"revision_date"`
	Tier          string    `db:"tier"`
	Rank          string    `db:"rank"`
	LeaguePoints  int       `db:"league_points"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// Match represents a stored game.
type Match struct {
	ID           int64     `db:"id"`
	MatchID      string    `db:"match_id"`
	Platform     string    `db:"platform"`
	QueueID      int       `db:"queue_id"`
	GameMode     string    `db:"game_mode"`
	GameDuration int64     `db:"game_duration"`
	GameVersion  string    `db:"game_version"`
	GameEndedAt  time.Time `db:"game_ended_at"`
	CreatedAt    time.Time `db:"created_at"`
}

// Participant represents a player's performance in a match.
type Participant struct {
	ID                   int64  `db:"id"`
	MatchID              int64  `db:"match_id"`
	PUUID                string `db:"puuid"`
	SummonerName         string `db:"summoner_name"`
	ChampionName         string `db:"champion_name"`
	TeamPosition         string `db:"team_position"`
	ChampionID           int    `db:"champion_id"`
	TeamID               int    `db:"team_id"`
	Kills                int    `db:"kills"`
	Deaths               int    `db:"deaths"`
	Assists              int    `db:"assists"`
	TotalMinionsKilled   int    `db:"total_minions_killed"`
	NeutralMinionsKilled int    `db:"neutral_minions_killed"`
	VisionScore          int    `db:"vision_score"`
	WardsPlaced          int    `db:"wards_placed"`
	WardsKilled          int    `db:"wards_killed"`
	DetectorWardsPlaced  int    `db:"detector_wards_placed"`
	DamageDealt          int    `db:"damage_dealt"`
	DamageTaken          int    `db:"damage_taken"`
	GoldEarned           int    `db:"gold_earned"`
	DragonKills          int    `db:"dragon_kills"`
	BaronKills           int    `db:"baron_kills"`
	TurretKills          int    `db:"turret_kills"`
	Win                  bool   `db:"win"`
	FirstBloodKill       bool   `db:"first_blood_kill"`
	FirstBloodAssist     bool   `db:"first_blood_assist"`
}

// SummonerMatch links a summoner to their tracked matches.
type SummonerMatch struct {
	SummonerID int64     `db:"summoner_id"`
	MatchID    int64     `db:"match_id"`
	CreatedAt  time.Time `db:"created_at"`
}

// MaxMatchesPerSummoner is the maximum number of matches tracked per summoner.
const MaxMatchesPerSummoner = 20
