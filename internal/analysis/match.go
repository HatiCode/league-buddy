package analysis

import (
	"errors"
	"fmt"

	"github.com/HatiCode/league-buddy/internal/models"
)

var (
	ErrParticipantNotFound = errors.New("participant not found in match")
	ErrMatchTooShort       = errors.New("match duration too short (likely a remake)")
)

const minMatchDurationSeconds = 60

func AnalyzeMatch(match *models.Match, puuid string) (*MatchAnalysis, error) {
	if match.Info.GameDuration < minMatchDurationSeconds {
		return nil, ErrMatchTooShort
	}

	participant, team, err := findParticipant(match, puuid)
	if err != nil {
		return nil, err
	}

	gameDurationMin := float64(match.Info.GameDuration) / 60.0

	metrics := MatchMetrics{
		MatchID:       match.Metadata.MatchID,
		ChampionName:  participant.ChampionName,
		Role:          participant.TeamPosition,
		GameDuration:  match.Info.GameDuration,
		Win:           participant.Win,
		TimeSpentDead: participant.TotalTimeSpentDead,
	}

	if participant.Challenges != nil {
		fillFromChallenges(&metrics, participant.Challenges, team)
	} else {
		fillFromRawStats(&metrics, participant, match)
	}

	fillComputedStats(&metrics, participant, match, gameDurationMin)

	return &MatchAnalysis{Metrics: metrics}, nil
}

func fillFromChallenges(metrics *MatchMetrics, challenges *models.Challenges, team *models.Team) {
	metrics.KDA = challenges.KDA
	metrics.KillParticipation = challenges.KillParticipation
	metrics.DamagePerMinute = challenges.DamagePerMinute
	metrics.DamageShare = challenges.TeamDamagePercentage
	metrics.GoldPerMinute = challenges.GoldPerMinute
	metrics.VisionScorePerMinute = challenges.VisionScorePerMinute
	metrics.DamageTakenShare = challenges.DamageTakenOnTeamPercentage
	metrics.HealShieldEffective = challenges.EffectiveHealAndShielding
	metrics.SoloKills = challenges.SoloKills
	metrics.ControlWardsPlaced = challenges.ControlWardsPlaced
	metrics.LaneMinionsFirst10Min = challenges.LaneMinionsFirst10Minutes
	metrics.EarlyLaningGoldExpAdvantage = challenges.EarlyLaningPhaseGoldExpAdvantage
	metrics.LaningGoldExpAdvantage = challenges.LaningPhaseGoldExpAdvantage
	metrics.MaxCsAdvantageOnLaneOpponent = challenges.MaxCsAdvantageOnLaneOpponent

	teamObjectiveKills := team.Objectives.Dragon.Kills +
		team.Objectives.Baron.Kills +
		team.Objectives.RiftHerald.Kills
	playerTakedowns := challenges.DragonTakedowns +
		challenges.BaronTakedowns +
		challenges.RiftHeraldTakedowns

	if teamObjectiveKills > 0 {
		metrics.ObjectiveParticipation = float64(playerTakedowns) / float64(teamObjectiveKills)
	}
}

func fillFromRawStats(metrics *MatchMetrics, participant *models.Participant, match *models.Match) {
	gameDurationMin := float64(match.Info.GameDuration) / 60.0
	deaths := participant.Deaths
	if deaths == 0 {
		deaths = 1
	}
	metrics.KDA = float64(participant.Kills+participant.Assists) / float64(deaths)
	metrics.DamagePerMinute = float64(participant.TotalDamageDealtToChampions) / gameDurationMin
	metrics.GoldPerMinute = float64(participant.GoldEarned) / gameDurationMin
	metrics.VisionScorePerMinute = float64(participant.VisionScore) / gameDurationMin
	metrics.ControlWardsPlaced = participant.DetectorWardsPlaced

	teamKills := sumTeamKills(match, participant.TeamID)
	if teamKills > 0 {
		metrics.KillParticipation = float64(participant.Kills+participant.Assists) / float64(teamKills)
	}

	teamDamage := sumTeamDamage(match, participant.TeamID)
	if teamDamage > 0 {
		metrics.DamageShare = float64(participant.TotalDamageDealtToChampions) / float64(teamDamage)
	}

	teamDamageTaken := sumTeamDamageTaken(match, participant.TeamID)
	if teamDamageTaken > 0 {
		metrics.DamageTakenShare = float64(participant.TotalDamageTaken) / float64(teamDamageTaken)
	}
}

func fillComputedStats(metrics *MatchMetrics, participant *models.Participant, match *models.Match, gameDurationMin float64) {
	totalCS := participant.TotalMinionsKilled + participant.NeutralMinionsKilled
	metrics.CSPerMinute = float64(totalCS) / gameDurationMin
	metrics.WardsPerMinute = float64(participant.WardsPlaced) / gameDurationMin
	metrics.CCPerMinute = float64(participant.TimeCCingOthers) / gameDurationMin
	metrics.DeathsPerMinute = float64(participant.Deaths) / gameDurationMin

	teamBuildingDamage := sumTeamBuildingDamage(match, participant.TeamID)
	if teamBuildingDamage > 0 {
		metrics.TurretDamageShare = float64(participant.DamageDealtToBuildings) / float64(teamBuildingDamage)
	}
}

func findParticipant(match *models.Match, puuid string) (*models.Participant, *models.Team, error) {
	var participant *models.Participant
	for i := range match.Info.Participants {
		if match.Info.Participants[i].PUUID == puuid {
			participant = &match.Info.Participants[i]
			break
		}
	}
	if participant == nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrParticipantNotFound, puuid)
	}

	var team *models.Team
	for i := range match.Info.Teams {
		if match.Info.Teams[i].TeamID == participant.TeamID {
			team = &match.Info.Teams[i]
			break
		}
	}

	return participant, team, nil
}

func sumTeamKills(match *models.Match, teamID int) int {
	total := 0
	for _, p := range match.Info.Participants {
		if p.TeamID == teamID {
			total += p.Kills
		}
	}
	return total
}

func sumTeamDamage(match *models.Match, teamID int) int {
	total := 0
	for _, p := range match.Info.Participants {
		if p.TeamID == teamID {
			total += p.TotalDamageDealtToChampions
		}
	}
	return total
}

func sumTeamDamageTaken(match *models.Match, teamID int) int {
	total := 0
	for _, p := range match.Info.Participants {
		if p.TeamID == teamID {
			total += p.TotalDamageTaken
		}
	}
	return total
}

func sumTeamBuildingDamage(match *models.Match, teamID int) int {
	total := 0
	for _, p := range match.Info.Participants {
		if p.TeamID == teamID {
			total += p.DamageDealtToBuildings
		}
	}
	return total
}
