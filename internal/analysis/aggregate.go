package analysis

import (
	"fmt"
	"math"
	"sort"
)

func AnalyzePlayer(params PlayerAnalysisParams) (*PlayerAnalysis, error) {
	if len(params.Matches) == 0 {
		return nil, fmt.Errorf("at least one match is required")
	}

	var analyses []MatchAnalysis
	for i := range params.Matches {
		match := &params.Matches[i]
		result, err := AnalyzeMatch(match, params.PUUID)
		if err != nil {
			continue
		}

		if params.Timelines != nil {
			if tl, ok := params.Timelines[match.Metadata.MatchID]; ok && tl != nil {
				lanePhase, err := AnalyzeLanePhase(tl, match, params.PUUID)
				if err == nil {
					result.LanePhase = lanePhase
				}
			}
		}

		analyses = append(analyses, *result)
	}

	if len(analyses) == 0 {
		return nil, fmt.Errorf("no valid matches to analyze (all may be remakes)")
	}

	analysis := &PlayerAnalysis{
		PUUID:        params.PUUID,
		GameName:     params.GameName,
		TagLine:      params.TagLine,
		TotalMatches: len(analyses),
		Matches:      analyses,
	}

	if params.League != nil {
		analysis.Tier = params.League.Tier
		analysis.Rank = params.League.Rank
		analysis.LeaguePoints = params.League.LeaguePoints
	}

	analysis.WinRate = computeWinRate(analyses)
	analysis.Averages = computeAverages(analyses)
	analysis.Consistency = computeConsistency(analyses)
	analysis.RoleBreakdown = computeRoleBreakdown(analyses)
	analysis.ChampionPool = computeChampionPool(analyses)
	analysis.Strengths, analysis.Weaknesses = identifyInsights(analysis.Averages, analysis.ChampionPool, analysis.Consistency)

	return analysis, nil
}

func computeWinRate(analyses []MatchAnalysis) float64 {
	wins := 0
	for _, a := range analyses {
		if a.Metrics.Win {
			wins++
		}
	}
	return float64(wins) / float64(len(analyses))
}

func computeAverages(analyses []MatchAnalysis) AverageMetrics {
	n := float64(len(analyses))
	var avg AverageMetrics

	for _, a := range analyses {
		m := a.Metrics
		avg.KDA += m.KDA
		avg.KillParticipation += m.KillParticipation
		avg.DamagePerMinute += m.DamagePerMinute
		avg.DamageShare += m.DamageShare
		avg.CSPerMinute += m.CSPerMinute
		avg.VisionScorePerMinute += m.VisionScorePerMinute
		avg.DeathsPerMinute += m.DeathsPerMinute
		avg.GoldPerMinute += m.GoldPerMinute
		avg.ObjectiveParticipation += m.ObjectiveParticipation
	}

	avg.KDA /= n
	avg.KillParticipation /= n
	avg.DamagePerMinute /= n
	avg.DamageShare /= n
	avg.CSPerMinute /= n
	avg.VisionScorePerMinute /= n
	avg.DeathsPerMinute /= n
	avg.GoldPerMinute /= n
	avg.ObjectiveParticipation /= n

	return avg
}

func computeConsistency(analyses []MatchAnalysis) ConsistencyMetrics {
	kdas := make([]float64, len(analyses))
	csPerMins := make([]float64, len(analyses))
	dpms := make([]float64, len(analyses))

	for i, a := range analyses {
		kdas[i] = a.Metrics.KDA
		csPerMins[i] = a.Metrics.CSPerMinute
		dpms[i] = a.Metrics.DamagePerMinute
	}

	return ConsistencyMetrics{
		KDAStdDev:      stddev(kdas),
		CSPerMinStdDev: stddev(csPerMins),
		DPMStdDev:      stddev(dpms),
	}
}

func computeRoleBreakdown(analyses []MatchAnalysis) []RoleStats {
	roleMap := make(map[string]*struct {
		games int
		wins  int
	})

	for _, a := range analyses {
		role := a.Metrics.Role
		if role == "" {
			role = "UNKNOWN"
		}
		entry, ok := roleMap[role]
		if !ok {
			entry = &struct {
				games int
				wins  int
			}{}
			roleMap[role] = entry
		}
		entry.games++
		if a.Metrics.Win {
			entry.wins++
		}
	}

	roles := make([]RoleStats, 0, len(roleMap))
	for role, entry := range roleMap {
		roles = append(roles, RoleStats{
			Role:        role,
			GamesPlayed: entry.games,
			WinRate:     float64(entry.wins) / float64(entry.games),
		})
	}

	sort.Slice(roles, func(i, j int) bool {
		return roles[i].GamesPlayed > roles[j].GamesPlayed
	})

	return roles
}

func computeChampionPool(analyses []MatchAnalysis) []ChampionStats {
	champMap := make(map[string]*struct {
		games  int
		wins   int
		kdaSum float64
	})

	for _, a := range analyses {
		name := a.Metrics.ChampionName
		entry, ok := champMap[name]
		if !ok {
			entry = &struct {
				games  int
				wins   int
				kdaSum float64
			}{}
			champMap[name] = entry
		}
		entry.games++
		entry.kdaSum += a.Metrics.KDA
		if a.Metrics.Win {
			entry.wins++
		}
	}

	pool := make([]ChampionStats, 0, len(champMap))
	for name, entry := range champMap {
		pool = append(pool, ChampionStats{
			ChampionName: name,
			GamesPlayed:  entry.games,
			WinRate:      float64(entry.wins) / float64(entry.games),
			AvgKDA:       entry.kdaSum / float64(entry.games),
		})
	}

	sort.Slice(pool, func(i, j int) bool {
		return pool[i].GamesPlayed > pool[j].GamesPlayed
	})

	return pool
}

type insightThreshold struct {
	category       string
	strengthMin    float64
	weaknessMax    float64
	getValue       func(AverageMetrics) float64
	strengthDesc   string
	weaknessDesc   string
	invertWeakness bool // true = high value is weakness (e.g. deaths)
}

var thresholds = []insightThreshold{
	{
		category: "combat", strengthMin: 3.0, weaknessMax: 1.5,
		getValue:     func(a AverageMetrics) float64 { return a.KDA },
		strengthDesc: "Strong KDA averaging %.1f -- effective at getting kills and staying alive",
		weaknessDesc: "Low KDA averaging %.1f -- dying too frequently relative to kill contribution",
	},
	{
		category: "combat", strengthMin: 0.65, weaknessMax: 0.40,
		getValue:     func(a AverageMetrics) float64 { return a.KillParticipation },
		strengthDesc: "High kill participation at %.0f%% -- consistently involved in team fights",
		weaknessDesc: "Low kill participation at %.0f%% -- missing team fights or playing too passively",
	},
	{
		category: "farming", strengthMin: 7.5, weaknessMax: 5.5,
		getValue:     func(a AverageMetrics) float64 { return a.CSPerMinute },
		strengthDesc: "Strong farming at %.1f CS/min -- efficient gold generation",
		weaknessDesc: "Low CS at %.1f per minute -- missing too much farm",
	},
	{
		category: "vision", strengthMin: 1.2, weaknessMax: 0.6,
		getValue:     func(a AverageMetrics) float64 { return a.VisionScorePerMinute },
		strengthDesc: "Excellent vision control at %.2f score/min",
		weaknessDesc: "Low vision score at %.2f per minute -- not warding enough",
	},
	{
		category: "combat", strengthMin: 0.28, weaknessMax: 0.15,
		getValue:     func(a AverageMetrics) float64 { return a.DamageShare },
		strengthDesc: "High team damage share at %.0f%% -- carrying damage output",
		weaknessDesc: "Low damage share at %.0f%% -- not contributing enough damage",
	},
	{
		category: "objectives", strengthMin: 0.60, weaknessMax: 0.30,
		getValue:     func(a AverageMetrics) float64 { return a.ObjectiveParticipation },
		strengthDesc: "Strong objective participation at %.0f%%",
		weaknessDesc: "Low objective participation at %.0f%% -- missing dragon and baron fights",
	},
	{
		category: "deaths", weaknessMax: 0.25, invertWeakness: true,
		getValue:     func(a AverageMetrics) float64 { return a.DeathsPerMinute },
		weaknessDesc: "High death rate at %.2f per minute -- positioning or decision-making needs work",
	},
}

func identifyInsights(avg AverageMetrics, championPool []ChampionStats, consistency ConsistencyMetrics) (strengths []Insight, weaknesses []Insight) {
	for _, t := range thresholds {
		value := t.getValue(avg)
		displayValue := value
		if t.category == "combat" && t.strengthMin == 0.65 || t.category == "combat" && t.strengthMin == 0.28 || t.category == "objectives" {
			displayValue = value * 100
		}

		if t.invertWeakness {
			if value >= t.weaknessMax {
				weaknesses = append(weaknesses, Insight{
					Category:    t.category,
					Description: fmt.Sprintf(t.weaknessDesc, displayValue),
					Value:       value,
				})
			}
			continue
		}

		if t.strengthMin > 0 && value >= t.strengthMin {
			strengths = append(strengths, Insight{
				Category:    t.category,
				Description: fmt.Sprintf(t.strengthDesc, displayValue),
				Value:       value,
				IsStrength:  true,
			})
		} else if value <= t.weaknessMax {
			weaknesses = append(weaknesses, Insight{
				Category:    t.category,
				Description: fmt.Sprintf(t.weaknessDesc, displayValue),
				Value:       value,
			})
		}
	}

	uniqueChamps := len(championPool)
	if uniqueChamps >= 8 {
		strengths = append(strengths, Insight{
			Category:    "champion_pool",
			Description: fmt.Sprintf("Deep champion pool with %d unique champions", uniqueChamps),
			Value:       float64(uniqueChamps),
			IsStrength:  true,
		})
	} else if uniqueChamps <= 2 {
		weaknesses = append(weaknesses, Insight{
			Category:    "champion_pool",
			Description: fmt.Sprintf("Narrow champion pool with only %d champion(s)", uniqueChamps),
			Value:       float64(uniqueChamps),
		})
	}

	if consistency.KDAStdDev < 1.0 {
		strengths = append(strengths, Insight{
			Category:    "consistency",
			Description: fmt.Sprintf("Very consistent KDA performance (stddev %.2f)", consistency.KDAStdDev),
			Value:       consistency.KDAStdDev,
			IsStrength:  true,
		})
	} else if consistency.KDAStdDev > 3.0 {
		weaknesses = append(weaknesses, Insight{
			Category:    "consistency",
			Description: fmt.Sprintf("Inconsistent performance with large KDA swings (stddev %.2f)", consistency.KDAStdDev),
			Value:       consistency.KDAStdDev,
		})
	}

	return strengths, weaknesses
}

func stddev(values []float64) float64 {
	n := float64(len(values))
	if n < 2 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / n

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	return math.Sqrt(variance / n)
}
