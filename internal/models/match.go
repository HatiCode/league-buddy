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
	Participants []string `json:"participants"`
}

// MatchInfo contains the actual game data.
type MatchInfo struct {
	EndOfGameResult    string        `json:"endOfGameResult"`
	GameMode           string        `json:"gameMode"`
	GameType           string        `json:"gameType"`
	GameVersion        string        `json:"gameVersion"`
	PlatformID         string        `json:"platformId"`
	GameCreation       int64         `json:"gameCreation"`
	GameDuration       int64         `json:"gameDuration"`
	GameEndTimestamp   int64         `json:"gameEndTimestamp"`
	GameID             int64         `json:"gameId"`
	GameStartTimestamp int64         `json:"gameStartTimestamp"`
	MapID              int           `json:"mapId"`
	QueueID            int           `json:"queueId"`
	Participants       []Participant `json:"participants"`
	Teams              []Team        `json:"teams"`
}

// Participant represents a player's performance in a match.
type Participant struct {
	PUUID                          string      `json:"puuid"`
	RiotIdGameName                 string      `json:"riotIdGameName"`
	RiotIdTagline                  string      `json:"riotIdTagline"`
	SummonerName                   string      `json:"summonerName"`
	ChampionName                   string      `json:"championName"`
	TeamPosition                   string      `json:"teamPosition"`
	Role                           string      `json:"role"`
	Lane                           string      `json:"lane"`
	ChampionID                     int         `json:"championId"`
	TeamID                         int         `json:"teamId"` // 100 = blue, 200 = red
	Kills                          int         `json:"kills"`
	Deaths                         int         `json:"deaths"`
	Assists                        int         `json:"assists"`
	TotalDamageDealtToChampions    int         `json:"totalDamageDealtToChampions"`
	MagicDamageDealtToChampions    int         `json:"magicDamageDealtToChampions"`
	PhysicalDamageDealtToChampions int         `json:"physicalDamageDealtToChampions"`
	TrueDamageDealtToChampions     int         `json:"trueDamageDealtToChampions"`
	TotalDamageTaken               int         `json:"totalDamageTaken"`
	DamageSelfMitigated            int         `json:"damageSelfMitigated"`
	LargestKillingSpree            int         `json:"largestKillingSpree"`
	KillingSprees                  int         `json:"killingSprees"`
	DoubleKills                    int         `json:"doubleKills"`
	TripleKills                    int         `json:"tripleKills"`
	QuadraKills                    int         `json:"quadraKills"`
	PentaKills                     int         `json:"pentaKills"`
	TotalDamageShieldedOnTeammates int         `json:"totalDamageShieldedOnTeammates"`
	TotalHealsOnTeammates          int         `json:"totalHealsOnTeammates"`
	TotalHeal                      int         `json:"totalHeal"`
	TimeCCingOthers                int         `json:"timeCCingOthers"`
	TotalTimeCCDealt               int         `json:"totalTimeCCDealt"`
	GoldEarned                     int         `json:"goldEarned"`
	GoldSpent                      int         `json:"goldSpent"`
	TotalMinionsKilled             int         `json:"totalMinionsKilled"`
	NeutralMinionsKilled           int         `json:"neutralMinionsKilled"`
	VisionScore                    int         `json:"visionScore"`
	WardsPlaced                    int         `json:"wardsPlaced"`
	WardsKilled                    int         `json:"wardsKilled"`
	DetectorWardsPlaced            int         `json:"detectorWardsPlaced"`
	VisionWardsBoughtInGame        int         `json:"visionWardsBoughtInGame"`
	SightWardsBoughtInGame         int         `json:"sightWardsBoughtInGame"`
	DamageDealtToObjectives        int         `json:"damageDealtToObjectives"`
	DamageDealtToBuildings         int         `json:"damageDealtToBuildings"`
	DragonKills                    int         `json:"dragonKills"`
	BaronKills                     int         `json:"baronKills"`
	TurretKills                    int         `json:"turretKills"`
	TurretTakedowns                int         `json:"turretTakedowns"`
	InhibitorKills                 int         `json:"inhibitorKills"`
	InhibitorTakedowns             int         `json:"inhibitorTakedowns"`
	ObjectivesStolen               int         `json:"objectivesStolen"`
	ChampLevel                     int         `json:"champLevel"`
	ChampExperience                int         `json:"champExperience"`
	TimePlayed                     int         `json:"timePlayed"`
	TotalTimeSpentDead             int         `json:"totalTimeSpentDead"`
	LongestTimeSpentLiving         int         `json:"longestTimeSpentLiving"`
	Summoner1Id                    int         `json:"summoner1Id"`
	Summoner2Id                    int         `json:"summoner2Id"`
	Summoner1Casts                 int         `json:"summoner1Casts"`
	Summoner2Casts                 int         `json:"summoner2Casts"`
	Spell1Casts                    int         `json:"spell1Casts"`
	Spell2Casts                    int         `json:"spell2Casts"`
	Spell3Casts                    int         `json:"spell3Casts"`
	Spell4Casts                    int         `json:"spell4Casts"`
	Challenges                     *Challenges `json:"challenges,omitempty"`
	Win                            bool        `json:"win"`
	FirstBloodKill                 bool        `json:"firstBloodKill"`
	FirstBloodAssist               bool        `json:"firstBloodAssist"`
	FirstTowerKill                 bool        `json:"firstTowerKill"`
	FirstTowerAssist               bool        `json:"firstTowerAssist"`
	GameEndedInSurrender           bool        `json:"gameEndedInSurrender"`
	GameEndedInEarlySurrender      bool        `json:"gameEndedInEarlySurrender"`
}

// Challenges contains pre-computed analytical metrics from Riot.
type Challenges struct {
	MaxCsAdvantageOnLaneOpponent     float64 `json:"maxCsAdvantageOnLaneOpponent"`
	KDA                              float64 `json:"kda"`
	KillParticipation                float64 `json:"killParticipation"`
	DamagePerMinute                  float64 `json:"damagePerMinute"`
	TeamDamagePercentage             float64 `json:"teamDamagePercentage"`
	DamageTakenOnTeamPercentage      float64 `json:"damageTakenOnTeamPercentage"`
	GoldPerMinute                    float64 `json:"goldPerMinute"`
	VisionScorePerMinute             float64 `json:"visionScorePerMinute"`
	EnemyJungleMonsterKills          float64 `json:"enemyJungleMonsterKills"`
	AlliedJungleMonsterKills         float64 `json:"alliedJungleMonsterKills"`
	EffectiveHealAndShielding        float64 `json:"effectiveHealAndShielding"`
	SoloKills                        int     `json:"soloKills"`
	LaneMinionsFirst10Minutes        int     `json:"laneMinionsFirst10Minutes"`
	MaxLevelLeadLaneOpponent         int     `json:"maxLevelLeadLaneOpponent"`
	EarlyLaningPhaseGoldExpAdvantage int     `json:"earlyLaningPhaseGoldExpAdvantage"`
	LaningPhaseGoldExpAdvantage      int     `json:"laningPhaseGoldExpAdvantage"`
	TurretPlatesTaken                int     `json:"turretPlatesTaken"`
	SkillshotsHit                    int     `json:"skillshotsHit"`
	SkillshotsDodged                 int     `json:"skillshotsDodged"`
	DeathsByEnemyChamps              int     `json:"deathsByEnemyChamps"`
	ControlWardsPlaced               int     `json:"controlWardsPlaced"`
	StealthWardsPlaced               int     `json:"stealthWardsPlaced"`
	BaronTakedowns                   int     `json:"baronTakedowns"`
	DragonTakedowns                  int     `json:"dragonTakedowns"`
	RiftHeraldTakedowns              int     `json:"riftHeraldTakedowns"`
	ScuttleCrabKills                 int     `json:"scuttleCrabKills"`
	SaveAllyFromDeath                int     `json:"saveAllyFromDeath"`
}

// Team represents one side's performance and objectives.
type Team struct {
	TeamID     int            `json:"teamId"`
	Bans       []Ban          `json:"bans"`
	Objectives TeamObjectives `json:"objectives"`
	Win        bool           `json:"win"`
}

// Ban represents a champion ban in draft.
type Ban struct {
	ChampionID int `json:"championId"`
	PickTurn   int `json:"pickTurn"`
}

// TeamObjectives tracks objective control for a team.
type TeamObjectives struct {
	Baron      ObjectiveStats `json:"baron"`
	Champion   ObjectiveStats `json:"champion"`
	Dragon     ObjectiveStats `json:"dragon"`
	Horde      ObjectiveStats `json:"horde"`
	Inhibitor  ObjectiveStats `json:"inhibitor"`
	RiftHerald ObjectiveStats `json:"riftHerald"`
	Tower      ObjectiveStats `json:"tower"`
}

// ObjectiveStats tracks kills and first take for an objective type.
type ObjectiveStats struct {
	Kills int  `json:"kills"`
	First bool `json:"first"`
}
