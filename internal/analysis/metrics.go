package analysis

import "github.com/HatiCode/league-buddy/internal/models"

// MatchMetrics holds computed metrics for a single player in a single match.
type MatchMetrics struct {
	MatchID      string `json:"matchId"`
	ChampionName string `json:"championName"`
	Role         string `json:"role"`

	KDA                          float64 `json:"kda"`
	KillParticipation            float64 `json:"killParticipation"`
	DamagePerMinute              float64 `json:"damagePerMinute"`
	DamageShare                  float64 `json:"damageShare"`
	CSPerMinute                  float64 `json:"csPerMinute"`
	VisionScorePerMinute         float64 `json:"visionScorePerMinute"`
	WardsPerMinute               float64 `json:"wardsPerMinute"`
	CCPerMinute                  float64 `json:"ccPerMinute"`
	DeathsPerMinute              float64 `json:"deathsPerMinute"`
	GoldPerMinute                float64 `json:"goldPerMinute"`
	ObjectiveParticipation       float64 `json:"objectiveParticipation"`
	TurretDamageShare            float64 `json:"turretDamageShare"`
	DamageTakenShare             float64 `json:"damageTakenShare"`
	HealShieldEffective          float64 `json:"healShieldEffective"`
	MaxCsAdvantageOnLaneOpponent float64 `json:"maxCsAdvantageOnLaneOpponent"`

	GameDuration                int64 `json:"gameDuration"`
	SoloKills                   int   `json:"soloKills"`
	ControlWardsPlaced          int   `json:"controlWardsPlaced"`
	LaneMinionsFirst10Min       int   `json:"laneMinionsFirst10Min"`
	EarlyLaningGoldExpAdvantage int   `json:"earlyLaningGoldExpAdvantage"`
	LaningGoldExpAdvantage      int   `json:"laningGoldExpAdvantage"`
	TimeSpentDead               int   `json:"timeSpentDead"`

	Win bool `json:"win"`
}

// LanePhaseMetrics holds timeline-derived early game data.
type LanePhaseMetrics struct {
	GoldDiffAt10   int `json:"goldDiffAt10"`
	GoldDiffAt15   int `json:"goldDiffAt15"`
	CSDiffAt10     int `json:"csDiffAt10"`
	GoldAt10       int `json:"goldAt10"`
	GoldAt15       int `json:"goldAt15"`
	CSAt10         int `json:"csAt10"`
	CSAt15         int `json:"csAt15"`
	XPAt10         int `json:"xpAt10"`
	DeathsBefore10 int `json:"deathsBefore10"`
}

// MatchAnalysis combines match metrics with optional lane phase data.
type MatchAnalysis struct {
	Metrics   MatchMetrics      `json:"metrics"`
	LanePhase *LanePhaseMetrics `json:"lanePhase,omitempty"`
}

// AverageMetrics holds mean values across all analyzed matches.
type AverageMetrics struct {
	KDA                    float64 `json:"kda"`
	KillParticipation      float64 `json:"killParticipation"`
	DamagePerMinute        float64 `json:"damagePerMinute"`
	DamageShare            float64 `json:"damageShare"`
	CSPerMinute            float64 `json:"csPerMinute"`
	VisionScorePerMinute   float64 `json:"visionScorePerMinute"`
	DeathsPerMinute        float64 `json:"deathsPerMinute"`
	GoldPerMinute          float64 `json:"goldPerMinute"`
	ObjectiveParticipation float64 `json:"objectiveParticipation"`
}

// ConsistencyMetrics tracks standard deviation of key metrics.
type ConsistencyMetrics struct {
	KDAStdDev      float64 `json:"kdaStdDev"`
	CSPerMinStdDev float64 `json:"csPerMinStdDev"`
	DPMStdDev      float64 `json:"dpmStdDev"`
}

// ChampionStats tracks per-champion aggregated performance.
type ChampionStats struct {
	ChampionName string  `json:"championName"`
	AvgKDA       float64 `json:"avgKda"`
	WinRate      float64 `json:"winRate"`
	GamesPlayed  int     `json:"gamesPlayed"`
}

// RoleStats tracks per-role aggregated performance.
type RoleStats struct {
	Role        string  `json:"role"`
	WinRate     float64 `json:"winRate"`
	GamesPlayed int     `json:"gamesPlayed"`
}

// Insight represents a single identified strength or weakness.
type Insight struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Value       float64 `json:"value"`
	IsStrength  bool    `json:"isStrength"`
}

// PlayerAnalysis is the final output combining all analysis for the coaching LLM.
type PlayerAnalysis struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
	Tier     string `json:"tier,omitempty"`
	Rank     string `json:"rank,omitempty"`

	WinRate      float64 `json:"winRate"`
	LeaguePoints int     `json:"leaguePoints,omitempty"`
	TotalMatches int     `json:"totalMatches"`

	Averages      AverageMetrics     `json:"averages"`
	Consistency   ConsistencyMetrics `json:"consistency"`
	RoleBreakdown []RoleStats        `json:"roleBreakdown"`
	ChampionPool  []ChampionStats    `json:"championPool"`
	Strengths     []Insight          `json:"strengths"`
	Weaknesses    []Insight          `json:"weaknesses"`
	Matches       []MatchAnalysis    `json:"matches"`
}

// PlayerAnalysisParams bundles all inputs for player analysis.
type PlayerAnalysisParams struct {
	PUUID     string
	GameName  string
	TagLine   string
	Matches   []models.Match
	Timelines map[string]*models.Timeline
	League    *models.LeagueEntry
}
