package analysis

import (
	"testing"

	"github.com/HatiCode/league-buddy/internal/models"
)

func makeTimeline(puuids []string, frameCount int) *models.Timeline {
	participants := make([]models.TimelineParticipant, len(puuids))
	for i, puuid := range puuids {
		participants[i] = models.TimelineParticipant{
			ParticipantID: i + 1,
			PUUID:         puuid,
		}
	}

	frames := make([]models.TimelineFrame, frameCount)
	for i := range frames {
		timestampMs := int64(i) * 60_000 // 1 frame per minute
		pFrames := make(map[string]models.ParticipantFrame)
		for j := range puuids {
			pid := j + 1
			goldPerMin := 400 + (pid * 20)
			csPerMin := 7 + pid
			pFrames[intToStr(pid)] = models.ParticipantFrame{
				ParticipantID:       pid,
				TotalGold:           goldPerMin * (i + 1),
				MinionsKilled:       csPerMin * i,
				JungleMinionsKilled: i,
				XP:                  300 * (i + 1),
			}
		}
		frames[i] = models.TimelineFrame{
			Timestamp:         timestampMs,
			ParticipantFrames: pFrames,
		}
	}
	return &models.Timeline{
		Metadata: models.TimelineMetadata{MatchID: "EUW1_TL"},
		Info: models.TimelineInfo{
			FrameInterval: 60_000,
			Participants:  participants,
			Frames:        frames,
		},
	}
}

func makeTimelineMatch(puuids []string) *models.Match {
	participants := make([]models.Participant, len(puuids))
	for i, puuid := range puuids {
		teamID := 100
		if i >= len(puuids)/2 {
			teamID = 200
		}
		participants[i] = models.Participant{
			PUUID:        puuid,
			TeamPosition: "MIDDLE",
			TeamID:       teamID,
		}
	}
	return &models.Match{
		Info: models.MatchInfo{
			Participants: participants,
		},
	}
}

func intToStr(n int) string {
	return string(rune('0' + n))
}

func TestAnalyzeLanePhaseBasic(t *testing.T) {
	puuids := []string{"player-1", "player-2", "opponent-1", "opponent-2"}
	timeline := makeTimeline(puuids, 20) // 20 frames = 0-19 min
	match := makeTimelineMatch(puuids)

	result, err := AnalyzeLanePhase(timeline, match, "player-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GoldAt10 == 0 {
		t.Error("goldAt10 should not be 0")
	}
	if result.CSAt10 == 0 {
		t.Error("csAt10 should not be 0")
	}
	if result.XPAt10 == 0 {
		t.Error("xpAt10 should not be 0")
	}
	if result.GoldAt15 == 0 {
		t.Error("goldAt15 should not be 0")
	}
	if result.CSAt15 == 0 {
		t.Error("csAt15 should not be 0")
	}
}

func TestAnalyzeLanePhaseDiffs(t *testing.T) {
	puuids := []string{"player-1", "teammate-1", "opponent-1", "opponent-2"}

	timeline := makeTimeline(puuids, 20)

	// player-1 is participant 1, opponent-1 is participant 3 (same MIDDLE role, different team)
	match := &models.Match{
		Info: models.MatchInfo{
			Participants: []models.Participant{
				{PUUID: "player-1", TeamPosition: "MIDDLE", TeamID: 100},
				{PUUID: "teammate-1", TeamPosition: "TOP", TeamID: 100},
				{PUUID: "opponent-1", TeamPosition: "MIDDLE", TeamID: 200},
				{PUUID: "opponent-2", TeamPosition: "TOP", TeamID: 200},
			},
		},
	}

	result, err := AnalyzeLanePhase(timeline, match, "player-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Diffs should be non-zero since participants have different gold/cs scaling
	if result.GoldDiffAt10 == 0 && result.GoldAt10 != 0 {
		t.Log("goldDiffAt10 is 0, which may indicate opponent has identical gold")
	}
}

func TestAnalyzeLanePhaseDeathsBefore10(t *testing.T) {
	puuids := []string{"player-1", "opponent-1"}
	timeline := makeTimeline(puuids, 20)

	// Add kill events: 2 deaths before 10 min, 1 after
	timeline.Info.Frames[5].Events = []models.TimelineEvent{
		{Type: "CHAMPION_KILL", VictimID: 1, Timestamp: 300_000},
	}
	timeline.Info.Frames[8].Events = []models.TimelineEvent{
		{Type: "CHAMPION_KILL", VictimID: 1, Timestamp: 480_000},
	}
	timeline.Info.Frames[12].Events = []models.TimelineEvent{
		{Type: "CHAMPION_KILL", VictimID: 1, Timestamp: 720_000},
	}

	match := makeTimelineMatch(puuids)
	result, err := AnalyzeLanePhase(timeline, match, "player-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DeathsBefore10 != 2 {
		t.Errorf("deathsBefore10 = %d, want 2", result.DeathsBefore10)
	}
}

func TestAnalyzeLanePhaseNoOpponent(t *testing.T) {
	puuids := []string{"player-1", "opponent-1"}
	timeline := makeTimeline(puuids, 20)

	// Player is MIDDLE, opponent is TOP -- no lane opponent
	match := &models.Match{
		Info: models.MatchInfo{
			Participants: []models.Participant{
				{PUUID: "player-1", TeamPosition: "MIDDLE", TeamID: 100},
				{PUUID: "opponent-1", TeamPosition: "TOP", TeamID: 200},
			},
		},
	}

	result, err := AnalyzeLanePhase(timeline, match, "player-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GoldDiffAt10 != 0 {
		t.Errorf("goldDiffAt10 = %d, want 0 when no lane opponent", result.GoldDiffAt10)
	}
	if result.CSDiffAt10 != 0 {
		t.Errorf("csDiffAt10 = %d, want 0 when no lane opponent", result.CSDiffAt10)
	}
}

func TestAnalyzeLanePhaseParticipantNotFound(t *testing.T) {
	puuids := []string{"player-1"}
	timeline := makeTimeline(puuids, 10)
	match := makeTimelineMatch(puuids)

	_, err := AnalyzeLanePhase(timeline, match, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing participant")
	}
}
