package models

// Summoner represents a League of Legends player identity.
type Summoner struct {
	PUUID         string `json:"puuid"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int64  `json:"summonerLevel"`
	ProfileIconID int    `json:"profileIconId"`
}

// LeagueEntry represents a player's ranked standing.
type LeagueEntry struct {
	LeagueID     string      `json:"leagueId"`
	PUUID        string      `json:"puuid"`
	QueueType    string      `json:"queueType"`
	Tier         string      `json:"tier"`
	Rank         string      `json:"rank"`
	LeaguePoints int         `json:"leaguePoints"`
	Wins         int         `json:"wins"`
	Losses       int         `json:"losses"`
	HotStreak    bool        `json:"hotStreak"`
	Veteran      bool        `json:"veteran"`
	FreshBlood   bool        `json:"freshBlood"`
	Inactive     bool        `json:"inactive"`
	MiniSeries   *MiniSeries `json:"miniSeries,omitempty"`
}

// MiniSeries represents a player's promotion series.
type MiniSeries struct {
	Progress string `json:"progress"`
	Losses   int    `json:"losses"`
	Target   int    `json:"target"`
	Wins     int    `json:"wins"`
}

// QueueType constants for ranked queues (league API).
const (
	QueueRankedSolo = "RANKED_SOLO_5x5"
	QueueRankedFlex = "RANKED_FLEX_SR"
)

// QueueIDRankedSolo is the match API queue ID for Ranked Solo/Duo.
const QueueIDRankedSolo = 420
