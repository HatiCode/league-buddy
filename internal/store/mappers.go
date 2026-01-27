package store

import (
	"time"

	"github.com/HatiCode/league-buddy/internal/models"
)

// SummonerFromAPI converts Riot API account and summoner responses to a store entity.
func SummonerFromAPI(account *models.Account, summoner *models.Summoner, platform string) *Summoner {
	return &Summoner{
		PUUID:         summoner.PUUID,
		GameName:      account.GameName,
		TagLine:       account.TagLine,
		Platform:      platform,
		ProfileIconID: summoner.ProfileIconID,
		SummonerLevel: summoner.SummonerLevel,
		RevisionDate:  summoner.RevisionDate,
	}
}

// ApplyLeagueEntry updates a summoner with ranked information.
func (s *Summoner) ApplyLeagueEntry(entry *models.LeagueEntry) {
	s.Tier = entry.Tier
	s.Rank = entry.Rank
	s.LeaguePoints = entry.LeaguePoints
}

// MatchFromAPI converts a Riot API match response to a store entity.
func MatchFromAPI(m *models.Match) *Match {
	return &Match{
		MatchID:      m.Metadata.MatchID,
		Platform:     m.Info.PlatformID,
		QueueID:      m.Info.QueueID,
		GameMode:     m.Info.GameMode,
		GameDuration: m.Info.GameDuration,
		GameVersion:  m.Info.GameVersion,
		GameEndedAt:  time.UnixMilli(m.Info.GameCreation + (m.Info.GameDuration * 1000)),
	}
}

// ParticipantsFromAPI extracts all participants from a match.
func ParticipantsFromAPI(m *models.Match) []Participant {
	participants := make([]Participant, 0, len(m.Info.Participants))

	for _, p := range m.Info.Participants {
		participants = append(participants, Participant{
			PUUID:                p.PUUID,
			SummonerName:         p.SummonerName,
			ChampionID:           p.ChampionID,
			ChampionName:         p.ChampionName,
			TeamID:               p.TeamID,
			TeamPosition:         p.TeamPosition,
			Win:                  p.Win,
			Kills:                p.Kills,
			Deaths:               p.Deaths,
			Assists:              p.Assists,
			TotalMinionsKilled:   p.TotalMinionsKilled,
			NeutralMinionsKilled: p.NeutralMinionsKilled,
			VisionScore:          p.VisionScore,
			WardsPlaced:          p.WardsPlaced,
			WardsKilled:          p.WardsKilled,
			DetectorWardsPlaced:  p.DetectorWardsPlaced,
			DamageDealt:          p.TotalDamageDealtToChampions,
			DamageTaken:          p.TotalDamageTaken,
			GoldEarned:           p.GoldEarned,
			DragonKills:          p.DragonKills,
			BaronKills:           p.BaronKills,
			TurretKills:          p.TurretKills,
			FirstBloodKill:       p.FirstBloodKill,
			FirstBloodAssist:     p.FirstBloodAssist,
		})
	}

	return participants
}
