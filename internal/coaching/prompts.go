package coaching

import (
	"fmt"
	"strings"

	"github.com/HatiCode/league-buddy/internal/analysis"
)

func BuildInitialSystemPrompt(a *analysis.PlayerAnalysis) string {
	var b strings.Builder

	b.WriteString("You are an expert League of Legends coach. Your role is to analyze player statistics and provide actionable, specific advice to help them improve and climb the ranked ladder.\n\n")

	writePlayerContext(&b, a)
	writeAverages(&b, a.Averages)
	writeConsistency(&b, a.Consistency)
	writeInsights(&b, "Strengths", a.Strengths)
	writeInsights(&b, "Weaknesses", a.Weaknesses)
	writeChampionPool(&b, a.ChampionPool)
	writeRoleBreakdown(&b, a.RoleBreakdown)
	writeMatchHistory(&b, a.Matches)

	b.WriteString("## Response Format\n")
	b.WriteString("1. Summary (2-3 sentences assessing the player overall)\n")
	b.WriteString("2. Top 3 action items ranked by impact on climbing\n")
	b.WriteString("3. Specific advice for each identified weakness\n")
	b.WriteString("4. Champion and role recommendations based on their pool and performance\n")

	return b.String()
}

func BuildFollowUpSystemPrompt(current *analysis.PlayerAnalysis, previous *analysis.PlayerAnalysis, previousAdvice string) string {
	var b strings.Builder

	b.WriteString("You are an expert League of Legends coach conducting a follow-up session. You previously coached this player and now have new match data to assess their progress.\n\n")

	writePlayerContext(&b, current)
	writeAverages(&b, current.Averages)
	writeConsistency(&b, current.Consistency)
	writeInsights(&b, "Current Strengths", current.Strengths)
	writeInsights(&b, "Current Weaknesses", current.Weaknesses)
	writeChampionPool(&b, current.ChampionPool)
	writeMatchHistory(&b, current.Matches)

	b.WriteString("## Previous Session\n\n")
	b.WriteString("### Previous Averages\n")
	writeAverages(&b, previous.Averages)

	b.WriteString("### Progress Since Last Session\n")
	writeDeltas(&b, previous.Averages, current.Averages)

	b.WriteString("### Previous Coaching Advice\n")
	b.WriteString(previousAdvice)
	b.WriteString("\n\n")

	b.WriteString("## Response Format\n")
	b.WriteString("1. Progress assessment: what improved and what didn't since last session\n")
	b.WriteString("2. Acknowledge specific improvements\n")
	b.WriteString("3. Persistent weaknesses that need continued focus\n")
	b.WriteString("4. Updated top 3 action items based on new data\n")
	b.WriteString("5. Adjusted champion and role recommendations\n")

	return b.String()
}

func BuildUserPrompt(isFollowUp bool) string {
	if isFollowUp {
		return "This is a follow-up coaching session. Compare my progress since the last session and provide updated advice. What did I improve on? What still needs work? What should I focus on next?"
	}
	return "Analyze my recent matches and provide coaching advice to help me climb ranked. Be specific and actionable."
}

func writePlayerContext(b *strings.Builder, a *analysis.PlayerAnalysis) {
	b.WriteString("## Player Profile\n")
	fmt.Fprintf(b, "- Riot ID: %s#%s\n", a.GameName, a.TagLine)
	if a.Tier != "" {
		fmt.Fprintf(b, "- Rank: %s %s (%d LP)\n", a.Tier, a.Rank, a.LeaguePoints)
	}
	fmt.Fprintf(b, "- Win Rate: %.0f%% across %d matches\n", a.WinRate*100, a.TotalMatches)
	b.WriteString("\n")
}

func writeAverages(b *strings.Builder, avg analysis.AverageMetrics) {
	b.WriteString("### Key Averages\n")
	fmt.Fprintf(b, "- KDA: %.2f\n", avg.KDA)
	fmt.Fprintf(b, "- Kill Participation: %.0f%%\n", avg.KillParticipation*100)
	fmt.Fprintf(b, "- CS/min: %.1f\n", avg.CSPerMinute)
	fmt.Fprintf(b, "- Damage/min: %.0f\n", avg.DamagePerMinute)
	fmt.Fprintf(b, "- Damage Share: %.0f%%\n", avg.DamageShare*100)
	fmt.Fprintf(b, "- Vision Score/min: %.2f\n", avg.VisionScorePerMinute)
	fmt.Fprintf(b, "- Deaths/min: %.2f\n", avg.DeathsPerMinute)
	fmt.Fprintf(b, "- Gold/min: %.0f\n", avg.GoldPerMinute)
	fmt.Fprintf(b, "- Objective Participation: %.0f%%\n", avg.ObjectiveParticipation*100)
	b.WriteString("\n")
}

func writeConsistency(b *strings.Builder, c analysis.ConsistencyMetrics) {
	b.WriteString("### Consistency\n")
	fmt.Fprintf(b, "- KDA StdDev: %.2f\n", c.KDAStdDev)
	fmt.Fprintf(b, "- CS/min StdDev: %.2f\n", c.CSPerMinStdDev)
	fmt.Fprintf(b, "- DPM StdDev: %.0f\n", c.DPMStdDev)
	b.WriteString("\n")
}

func writeInsights(b *strings.Builder, label string, insights []analysis.Insight) {
	if len(insights) == 0 {
		return
	}
	fmt.Fprintf(b, "### %s\n", label)
	for _, insight := range insights {
		fmt.Fprintf(b, "- [%s] %s\n", insight.Category, insight.Description)
	}
	b.WriteString("\n")
}

func writeChampionPool(b *strings.Builder, pool []analysis.ChampionStats) {
	if len(pool) == 0 {
		return
	}
	b.WriteString("### Champion Pool\n")
	for _, c := range pool {
		fmt.Fprintf(b, "- %s: %d games, %.0f%% WR, %.2f avg KDA\n",
			c.ChampionName, c.GamesPlayed, c.WinRate*100, c.AvgKDA)
	}
	b.WriteString("\n")
}

func writeRoleBreakdown(b *strings.Builder, roles []analysis.RoleStats) {
	if len(roles) == 0 {
		return
	}
	b.WriteString("### Role Breakdown\n")
	for _, r := range roles {
		fmt.Fprintf(b, "- %s: %d games, %.0f%% WR\n", r.Role, r.GamesPlayed, r.WinRate*100)
	}
	b.WriteString("\n")
}

func writeMatchHistory(b *strings.Builder, matches []analysis.MatchAnalysis) {
	if len(matches) == 0 {
		return
	}
	b.WriteString("### Recent Matches\n")
	for _, m := range matches {
		result := "Loss"
		if m.Metrics.Win {
			result = "Win"
		}
		fmt.Fprintf(b, "- %s %s (%s): %.1f KDA, %.1f CS/min, %.0f DPM [%s]\n",
			m.Metrics.ChampionName, m.Metrics.Role, result,
			m.Metrics.KDA, m.Metrics.CSPerMinute, m.Metrics.DamagePerMinute,
			m.Metrics.MatchID)
	}
	b.WriteString("\n")
}

func writeDeltas(b *strings.Builder, prev, curr analysis.AverageMetrics) {
	writeDelta(b, "KDA", prev.KDA, curr.KDA, "%.2f", false)
	writeDelta(b, "Kill Participation", prev.KillParticipation*100, curr.KillParticipation*100, "%.0f%%", false)
	writeDelta(b, "CS/min", prev.CSPerMinute, curr.CSPerMinute, "%.1f", false)
	writeDelta(b, "Damage/min", prev.DamagePerMinute, curr.DamagePerMinute, "%.0f", false)
	writeDelta(b, "Damage Share", prev.DamageShare*100, curr.DamageShare*100, "%.0f%%", false)
	writeDelta(b, "Vision Score/min", prev.VisionScorePerMinute, curr.VisionScorePerMinute, "%.2f", false)
	writeDelta(b, "Deaths/min", prev.DeathsPerMinute, curr.DeathsPerMinute, "%.2f", true)
	writeDelta(b, "Gold/min", prev.GoldPerMinute, curr.GoldPerMinute, "%.0f", false)
	writeDelta(b, "Objective Participation", prev.ObjectiveParticipation*100, curr.ObjectiveParticipation*100, "%.0f%%", false)
	b.WriteString("\n")
}

func writeDelta(b *strings.Builder, name string, prev, curr float64, format string, lowerIsBetter bool) {
	diff := curr - prev
	direction := "unchanged"
	if lowerIsBetter {
		diff = -diff
	}
	if diff > 0.01 {
		direction = "improved"
	} else if diff < -0.01 {
		direction = "regressed"
	}
	prevStr := fmt.Sprintf(format, prev)
	currStr := fmt.Sprintf(format, curr)
	fmt.Fprintf(b, "- %s: %s -> %s (%s)\n", name, prevStr, currStr, direction)
}
