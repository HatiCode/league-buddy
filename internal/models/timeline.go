package models

// Timeline represents match timeline data from Riot API.
type Timeline struct {
	Metadata TimelineMetadata `json:"metadata"`
	Info     TimelineInfo     `json:"info"`
}

// TimelineMetadata contains timeline metadata.
type TimelineMetadata struct {
	DataVersion  string   `json:"dataVersion"`
	MatchID      string   `json:"matchId"`
	Participants []string `json:"participants"`
}

// TimelineInfo contains the timeline frames and participant info.
type TimelineInfo struct {
	EndOfGameResult string                `json:"endOfGameResult"`
	FrameInterval   int64                 `json:"frameInterval"`
	GameID          int64                 `json:"gameId"`
	Participants    []TimelineParticipant `json:"participants"`
	Frames          []TimelineFrame       `json:"frames"`
}

// TimelineParticipant maps participant ID to PUUID.
type TimelineParticipant struct {
	ParticipantID int    `json:"participantId"`
	PUUID         string `json:"puuid"`
}

// TimelineFrame represents a snapshot of game state at a specific time.
type TimelineFrame struct {
	Events            []TimelineEvent           `json:"events"`
	ParticipantFrames map[string]ParticipantFrame `json:"participantFrames"`
	Timestamp         int64                     `json:"timestamp"`
}

// TimelineEvent represents a game event.
type TimelineEvent struct {
	Timestamp     int64  `json:"timestamp"`
	RealTimestamp int64  `json:"realTimestamp"`
	Type          string `json:"type"`

	// Optional fields depending on event type
	ParticipantID       int    `json:"participantId,omitempty"`
	KillerID            int    `json:"killerId,omitempty"`
	VictimID            int    `json:"victimId,omitempty"`
	AssistingParticipantIDs []int `json:"assistingParticipantIds,omitempty"`
	Position            *Position `json:"position,omitempty"`
	ItemID              int    `json:"itemId,omitempty"`
	SkillSlot           int    `json:"skillSlot,omitempty"`
	LevelUpType         string `json:"levelUpType,omitempty"`
	WardType            string `json:"wardType,omitempty"`
	CreatorID           int    `json:"creatorId,omitempty"`
	BuildingType        string `json:"buildingType,omitempty"`
	TowerType           string `json:"towerType,omitempty"`
	LaneType            string `json:"laneType,omitempty"`
	TeamID              int    `json:"teamId,omitempty"`
	MonsterType         string `json:"monsterType,omitempty"`
	MonsterSubType      string `json:"monsterSubType,omitempty"`
	KillerTeamID        int    `json:"killerTeamId,omitempty"`
	Bounty              int    `json:"bounty,omitempty"`
	KillStreakLength    int    `json:"killStreakLength,omitempty"`
}

// ParticipantFrame represents a participant's state at a frame.
type ParticipantFrame struct {
	ChampionStats             ChampionStats `json:"championStats"`
	CurrentGold               int           `json:"currentGold"`
	DamageStats               DamageStats   `json:"damageStats"`
	GoldPerSecond             int           `json:"goldPerSecond"`
	JungleMinionsKilled       int           `json:"jungleMinionsKilled"`
	Level                     int           `json:"level"`
	MinionsKilled             int           `json:"minionsKilled"`
	ParticipantID             int           `json:"participantId"`
	Position                  Position      `json:"position"`
	TimeEnemySpentControlled  int           `json:"timeEnemySpentControlled"`
	TotalGold                 int           `json:"totalGold"`
	XP                        int           `json:"xp"`
}

// ChampionStats represents champion statistics at a frame.
type ChampionStats struct {
	AbilityHaste         int `json:"abilityHaste"`
	AbilityPower         int `json:"abilityPower"`
	Armor                int `json:"armor"`
	ArmorPen             int `json:"armorPen"`
	ArmorPenPercent      int `json:"armorPenPercent"`
	AttackDamage         int `json:"attackDamage"`
	AttackSpeed          int `json:"attackSpeed"`
	BonusArmorPenPercent int `json:"bonusArmorPenPercent"`
	BonusMagicPenPercent int `json:"bonusMagicPenPercent"`
	CCReduction          int `json:"ccReduction"`
	CooldownReduction    int `json:"cooldownReduction"`
	Health               int `json:"health"`
	HealthMax            int `json:"healthMax"`
	HealthRegen          int `json:"healthRegen"`
	Lifesteal            int `json:"lifesteal"`
	MagicPen             int `json:"magicPen"`
	MagicPenPercent      int `json:"magicPenPercent"`
	MagicResist          int `json:"magicResist"`
	MovementSpeed        int `json:"movementSpeed"`
	Omnivamp             int `json:"omnivamp"`
	PhysicalVamp         int `json:"physicalVamp"`
	Power                int `json:"power"`
	PowerMax             int `json:"powerMax"`
	PowerRegen           int `json:"powerRegen"`
	SpellVamp            int `json:"spellVamp"`
}

// DamageStats represents damage statistics at a frame.
type DamageStats struct {
	MagicDamageDone              int `json:"magicDamageDone"`
	MagicDamageDoneToChampions   int `json:"magicDamageDoneToChampions"`
	MagicDamageTaken             int `json:"magicDamageTaken"`
	PhysicalDamageDone           int `json:"physicalDamageDone"`
	PhysicalDamageDoneToChampions int `json:"physicalDamageDoneToChampions"`
	PhysicalDamageTaken          int `json:"physicalDamageTaken"`
	TotalDamageDone              int `json:"totalDamageDone"`
	TotalDamageDoneToChampions   int `json:"totalDamageDoneToChampions"`
	TotalDamageTaken             int `json:"totalDamageTaken"`
	TrueDamageDone               int `json:"trueDamageDone"`
	TrueDamageDoneToChampions    int `json:"trueDamageDoneToChampions"`
	TrueDamageTaken              int `json:"trueDamageTaken"`
}

// Position represents a coordinate on the map.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
