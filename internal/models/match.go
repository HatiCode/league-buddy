package models

// Match represents a completed game with full details.
type Match struct {
	Metadata MatchMetadata `json:"metadata"`
	Info     MatchInfo     `json:"info"`
}

// MatchMetadata contains match identification data.
type MatchMetadata struct {
	MatchID      string   `json:"matchId"`
	DataVersion  string   `json:"dataVersion"`
	Participants []string `json:"participants"` // List of PUUIDs
}

// MatchInfo contains the actual game data.
type MatchInfo struct {
	GameCreation     int64         `json:"gameCreation"`     // Unix timestamp in ms
	GameDuration     int64         `json:"gameDuration"`     // Duration in seconds
	GameID           int64         `json:"gameId"`
	GameMode         string        `json:"gameMode"`
	GameType         string        `json:"gameType"`
	GameVersion      string        `json:"gameVersion"`
	MapID            int           `json:"mapId"`
	QueueID          int           `json:"queueId"`
	PlatformID       string        `json:"platformId"`
	Participants     []Participant `json:"participants"`
	Teams            []Team        `json:"teams"`
}

// Participant represents a player's performance in a match.
type Participant struct {
	PUUID                      string `json:"puuid"`
	SummonerID                 string `json:"summonerId"`
	SummonerName               string `json:"summonerName"`
	ChampionID                 int    `json:"championId"`
	ChampionName               string `json:"championName"`
	TeamID                     int    `json:"teamId"` // 100 = blue, 200 = red
	TeamPosition               string `json:"teamPosition"`
	Role                       string `json:"role"`
	Lane                       string `json:"lane"`
	Win                        bool   `json:"win"`

	// Combat stats
	Kills                      int    `json:"kills"`
	Deaths                     int    `json:"deaths"`
	Assists                    int    `json:"assists"`
	TotalDamageDealtToChampions int   `json:"totalDamageDealtToChampions"`
	TotalDamageTaken           int    `json:"totalDamageTaken"`
	LargestKillingSpree        int    `json:"largestKillingSpree"`
	DoubleKills                int    `json:"doubleKills"`
	TripleKills                int    `json:"tripleKills"`
	QuadraKills                int    `json:"quadraKills"`
	PentaKills                 int    `json:"pentaKills"`
	FirstBloodKill             bool   `json:"firstBloodKill"`
	FirstBloodAssist           bool   `json:"firstBloodAssist"`

	// Economy stats
	GoldEarned                 int    `json:"goldEarned"`
	GoldSpent                  int    `json:"goldSpent"`
	TotalMinionsKilled         int    `json:"totalMinionsKilled"`
	NeutralMinionsKilled       int    `json:"neutralMinionsKilled"`

	// Vision stats
	VisionScore                int    `json:"visionScore"`
	WardsPlaced                int    `json:"wardsPlaced"`
	WardsKilled                int    `json:"wardsKilled"`
	DetectorWardsPlaced        int    `json:"detectorWardsPlaced"`

	// Objective stats
	DragonKills                int    `json:"dragonKills"`
	BaronKills                 int    `json:"baronKills"`
	TurretKills                int    `json:"turretKills"`
	InhibitorKills             int    `json:"inhibitorKills"`
	ObjectivesStolen           int    `json:"objectivesStolen"`

	// Timeline
	ChampLevel                 int    `json:"champLevel"`
	TimePlayed                 int    `json:"timePlayed"`
	TotalTimeSpentDead         int    `json:"totalTimeSpentDead"`
}

// Team represents one side's performance and objectives.
type Team struct {
	TeamID     int              `json:"teamId"`
	Win        bool             `json:"win"`
	Objectives TeamObjectives   `json:"objectives"`
}

// TeamObjectives tracks objective control for a team.
type TeamObjectives struct {
	Baron      ObjectiveStats `json:"baron"`
	Champion   ObjectiveStats `json:"champion"`
	Dragon     ObjectiveStats `json:"dragon"`
	Inhibitor  ObjectiveStats `json:"inhibitor"`
	RiftHerald ObjectiveStats `json:"riftHerald"`
	Tower      ObjectiveStats `json:"tower"`
}

// ObjectiveStats tracks kills and first take for an objective type.
type ObjectiveStats struct {
	First bool `json:"first"`
	Kills int  `json:"kills"`
}
