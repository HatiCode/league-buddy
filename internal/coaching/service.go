package coaching

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HatiCode/league-buddy/internal/analysis"
	"github.com/HatiCode/league-buddy/internal/store"
)

// TrendPoint represents a single coaching session as a data point for progress tracking.
type TrendPoint struct {
	SessionDate time.Time              `json:"sessionDate"`
	MatchCount  int                    `json:"matchCount"`
	WinRate     float64                `json:"winRate"`
	Tier        string                 `json:"tier,omitempty"`
	Rank        string                 `json:"rank,omitempty"`
	Averages    analysis.AverageMetrics `json:"averages"`
}

// PlayerProgress holds the full trend data across all coaching sessions.
type PlayerProgress struct {
	PUUID    string       `json:"puuid"`
	GameName string       `json:"gameName"`
	TagLine  string       `json:"tagLine"`
	Sessions int          `json:"sessions"`
	Trend    []TrendPoint `json:"trend"`
}

// CoachingResponse holds the result of a coaching session.
type CoachingResponse struct {
	Advice     string `json:"advice"`
	IsFollowUp bool   `json:"isFollowUp"`
	NewMatches int    `json:"newMatches"`
}

// Service orchestrates the coaching flow: prompt building, LLM calls, and session persistence.
type Service struct {
	llm   LLMClient
	store store.CoachingSessionRepository
}

// NewService creates a coaching service. Pass nil for st to disable session persistence.
func NewService(llm LLMClient, st store.CoachingSessionRepository) *Service {
	return &Service{
		llm:   llm,
		store: st,
	}
}

// Coach runs a coaching session for the given player analysis.
// matchIDs should be ordered most-recent-first; matchIDs[0] is used as the session watermark.
func (s *Service) Coach(ctx context.Context, playerAnalysis *analysis.PlayerAnalysis, matchIDs []string) (*CoachingResponse, error) {
	var previousSession *store.CoachingSession
	if s.store != nil {
		var err error
		previousSession, err = s.store.GetLatestCoachingSession(ctx, playerAnalysis.PUUID)
		if err != nil {
			return nil, fmt.Errorf("get previous session: %w", err)
		}
	}

	system, isFollowUp, err := s.buildPrompts(playerAnalysis, previousSession)
	if err != nil {
		return nil, err
	}

	user := BuildUserPrompt(isFollowUp)

	advice, err := s.llm.Complete(ctx, system, user)
	if err != nil {
		return nil, fmt.Errorf("llm complete: %w", err)
	}

	if s.store != nil {
		if err := s.saveSession(ctx, playerAnalysis, matchIDs, advice); err != nil {
			return nil, fmt.Errorf("save session: %w", err)
		}
	}

	return &CoachingResponse{
		Advice:     advice,
		IsFollowUp: isFollowUp,
		NewMatches: len(matchIDs),
	}, nil
}

func (s *Service) buildPrompts(current *analysis.PlayerAnalysis, previous *store.CoachingSession) (string, bool, error) {
	if previous == nil {
		return BuildInitialSystemPrompt(current), false, nil
	}

	var previousAnalysis analysis.PlayerAnalysis
	if err := json.Unmarshal(previous.Analysis, &previousAnalysis); err != nil {
		return "", false, fmt.Errorf("unmarshal previous analysis: %w", err)
	}

	system := BuildFollowUpSystemPrompt(current, &previousAnalysis, previous.Advice)
	return system, true, nil
}

func (s *Service) saveSession(ctx context.Context, playerAnalysis *analysis.PlayerAnalysis, matchIDs []string, advice string) error {
	analysisJSON, err := json.Marshal(playerAnalysis)
	if err != nil {
		return fmt.Errorf("marshal analysis: %w", err)
	}

	matchIDsJSON, err := json.Marshal(matchIDs)
	if err != nil {
		return fmt.Errorf("marshal match IDs: %w", err)
	}

	latestMatchID := ""
	if len(matchIDs) > 0 {
		latestMatchID = matchIDs[0]
	}

	session := &store.CoachingSession{
		PUUID:         playerAnalysis.PUUID,
		LatestMatchID: latestMatchID,
		MatchIDs:      matchIDsJSON,
		Analysis:      analysisJSON,
		Advice:        advice,
	}

	return s.store.SaveCoachingSession(ctx, session)
}

// GetProgress loads all coaching sessions for a player and returns trend data.
func (s *Service) GetProgress(ctx context.Context, puuid string) (*PlayerProgress, error) {
	if s.store == nil {
		return nil, fmt.Errorf("database is required for progress tracking")
	}

	sessions, err := s.store.GetCoachingSessions(ctx, puuid)
	if err != nil {
		return nil, fmt.Errorf("get coaching sessions: %w", err)
	}

	progress := &PlayerProgress{
		PUUID:    puuid,
		Sessions: len(sessions),
		Trend:    make([]TrendPoint, 0, len(sessions)),
	}

	for i := range sessions {
		var pa analysis.PlayerAnalysis
		if err := json.Unmarshal(sessions[i].Analysis, &pa); err != nil {
			continue
		}

		if i == 0 {
			progress.GameName = pa.GameName
			progress.TagLine = pa.TagLine
		}

		var matchIDs []string
		_ = json.Unmarshal(sessions[i].MatchIDs, &matchIDs)

		progress.Trend = append(progress.Trend, TrendPoint{
			SessionDate: sessions[i].CreatedAt,
			MatchCount:  len(matchIDs),
			WinRate:     pa.WinRate,
			Tier:        pa.Tier,
			Rank:        pa.Rank,
			Averages:    pa.Averages,
		})
	}

	return progress, nil
}
