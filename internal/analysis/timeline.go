package analysis

import (
	"fmt"
	"strconv"

	"github.com/HatiCode/league-buddy/internal/models"
)

const (
	tenMinutesMs     = 600_000
	fifteenMinutesMs = 900_000
)

func AnalyzeLanePhase(timeline *models.Timeline, match *models.Match, puuid string) (*LanePhaseMetrics, error) {
	participantID, err := findTimelineParticipantID(timeline, puuid)
	if err != nil {
		return nil, err
	}

	opponentID := findLaneOpponent(match, puuid)
	playerKey := strconv.Itoa(participantID)
	opponentKey := strconv.Itoa(opponentID)

	metrics := &LanePhaseMetrics{}

	frame10 := findFrameAtTime(timeline.Info.Frames, tenMinutesMs)
	if frame10 != nil {
		if pf, ok := frame10.ParticipantFrames[playerKey]; ok {
			metrics.GoldAt10 = pf.TotalGold
			metrics.CSAt10 = pf.MinionsKilled + pf.JungleMinionsKilled
			metrics.XPAt10 = pf.XP

			if opponentID > 0 {
				if of, ok := frame10.ParticipantFrames[opponentKey]; ok {
					metrics.GoldDiffAt10 = metrics.GoldAt10 - of.TotalGold
					metrics.CSDiffAt10 = metrics.CSAt10 - (of.MinionsKilled + of.JungleMinionsKilled)
				}
			}
		}
	}

	frame15 := findFrameAtTime(timeline.Info.Frames, fifteenMinutesMs)
	if frame15 != nil {
		if pf, ok := frame15.ParticipantFrames[playerKey]; ok {
			metrics.GoldAt15 = pf.TotalGold
			metrics.CSAt15 = pf.MinionsKilled + pf.JungleMinionsKilled

			if opponentID > 0 {
				if of, ok := frame15.ParticipantFrames[opponentKey]; ok {
					metrics.GoldDiffAt15 = metrics.GoldAt15 - of.TotalGold
				}
			}
		}
	}

	metrics.DeathsBefore10 = countDeathsBefore(timeline.Info.Frames, participantID, tenMinutesMs)

	return metrics, nil
}

func findTimelineParticipantID(timeline *models.Timeline, puuid string) (int, error) {
	for _, p := range timeline.Info.Participants {
		if p.PUUID == puuid {
			return p.ParticipantID, nil
		}
	}
	return 0, fmt.Errorf("%w: %s", ErrParticipantNotFound, puuid)
}

func findLaneOpponent(match *models.Match, puuid string) int {
	var playerRole string
	var playerTeamID int
	for _, p := range match.Info.Participants {
		if p.PUUID == puuid {
			playerRole = p.TeamPosition
			playerTeamID = p.TeamID
			break
		}
	}

	if playerRole == "" {
		return 0
	}

	for i, p := range match.Info.Participants {
		if p.TeamID != playerTeamID && p.TeamPosition == playerRole {
			return i + 1 // participant IDs are 1-indexed
		}
	}
	return 0
}

func findFrameAtTime(frames []models.TimelineFrame, targetMs int64) *models.TimelineFrame {
	var closest *models.TimelineFrame
	closestDist := int64(1<<63 - 1)

	for i := range frames {
		dist := frames[i].Timestamp - targetMs
		if dist < 0 {
			dist = -dist
		}
		if dist < closestDist {
			closestDist = dist
			closest = &frames[i]
		}
	}
	return closest
}

func countDeathsBefore(frames []models.TimelineFrame, participantID int, beforeMs int64) int {
	deaths := 0
	for _, frame := range frames {
		for _, event := range frame.Events {
			if event.Type == "CHAMPION_KILL" && event.VictimID == participantID && event.Timestamp < beforeMs {
				deaths++
			}
		}
	}
	return deaths
}
