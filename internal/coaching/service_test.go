package coaching

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/HatiCode/league-buddy/internal/analysis"
	"github.com/HatiCode/league-buddy/internal/store"
)

type mockLLM struct {
	response string
	err      error
	system   string
	user     string
}

func (m *mockLLM) Complete(_ context.Context, system string, user string) (string, error) {
	m.system = system
	m.user = user
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

type mockSessionStore struct {
	latestSession *store.CoachingSession
	sessions      []store.CoachingSession
	savedSession  *store.CoachingSession
	getErr        error
	sessionsErr   error
	saveErr       error
}

func (m *mockSessionStore) GetLatestCoachingSession(_ context.Context, _ string) (*store.CoachingSession, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.latestSession, nil
}

func (m *mockSessionStore) GetCoachingSessions(_ context.Context, _ string) ([]store.CoachingSession, error) {
	if m.sessionsErr != nil {
		return nil, m.sessionsErr
	}
	return m.sessions, nil
}

func (m *mockSessionStore) SaveCoachingSession(_ context.Context, session *store.CoachingSession) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedSession = session
	return nil
}

func TestCoachInitialSession(t *testing.T) {
	llm := &mockLLM{response: "Here is your coaching advice."}
	st := &mockSessionStore{}
	svc := NewService(llm, st)

	a := makeTestAnalysis()
	matchIDs := []string{"EUW1_001", "EUW1_002"}

	resp, err := svc.Coach(context.Background(), a, matchIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsFollowUp {
		t.Error("expected initial session, got follow-up")
	}
	if resp.Advice != "Here is your coaching advice." {
		t.Errorf("advice = %q, want %q", resp.Advice, "Here is your coaching advice.")
	}
	if resp.NewMatches != 2 {
		t.Errorf("newMatches = %d, want 2", resp.NewMatches)
	}

	if !strings.Contains(llm.system, "League of Legends coach") {
		t.Error("system prompt missing coaching persona")
	}
	if strings.Contains(llm.system, "follow-up") {
		t.Error("system prompt should not contain follow-up for initial session")
	}

	if st.savedSession == nil {
		t.Fatal("expected session to be saved")
	}
	if st.savedSession.PUUID != "test-puuid" {
		t.Errorf("saved puuid = %q, want %q", st.savedSession.PUUID, "test-puuid")
	}
	if st.savedSession.LatestMatchID != "EUW1_001" {
		t.Errorf("saved latestMatchID = %q, want %q", st.savedSession.LatestMatchID, "EUW1_001")
	}
	if st.savedSession.Advice != "Here is your coaching advice." {
		t.Errorf("saved advice = %q, want %q", st.savedSession.Advice, "Here is your coaching advice.")
	}
}

func TestCoachFollowUpSession(t *testing.T) {
	previousAnalysis := makeTestAnalysis()
	previousAnalysis.Averages.KDA = 2.5
	analysisJSON, _ := json.Marshal(previousAnalysis)
	matchIDsJSON, _ := json.Marshal([]string{"EUW1_000"})

	llm := &mockLLM{response: "Updated coaching advice."}
	st := &mockSessionStore{
		latestSession: &store.CoachingSession{
			PUUID:         "test-puuid",
			LatestMatchID: "EUW1_000",
			MatchIDs:      matchIDsJSON,
			Analysis:      analysisJSON,
			Advice:        "Previous advice.",
		},
	}
	svc := NewService(llm, st)

	a := makeTestAnalysis()
	matchIDs := []string{"EUW1_001", "EUW1_002"}

	resp, err := svc.Coach(context.Background(), a, matchIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsFollowUp {
		t.Error("expected follow-up session, got initial")
	}
	if resp.Advice != "Updated coaching advice." {
		t.Errorf("advice = %q, want %q", resp.Advice, "Updated coaching advice.")
	}

	if !strings.Contains(llm.system, "follow-up") {
		t.Error("system prompt missing follow-up indicator")
	}
	if !strings.Contains(llm.system, "Previous advice.") {
		t.Error("system prompt missing previous advice")
	}
	if !strings.Contains(llm.system, "KDA: 2.50 -> 3.50 (improved)") {
		t.Error("system prompt missing KDA delta")
	}
}

func TestCoachNilStore(t *testing.T) {
	llm := &mockLLM{response: "Advice without store."}
	svc := NewService(llm, nil)

	a := makeTestAnalysis()
	matchIDs := []string{"EUW1_001"}

	resp, err := svc.Coach(context.Background(), a, matchIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsFollowUp {
		t.Error("expected initial session when store is nil")
	}
	if resp.Advice != "Advice without store." {
		t.Errorf("advice = %q, want %q", resp.Advice, "Advice without store.")
	}
}

func TestCoachLLMError(t *testing.T) {
	llm := &mockLLM{err: errors.New("rate limited")}
	svc := NewService(llm, nil)

	a := makeTestAnalysis()
	_, err := svc.Coach(context.Background(), a, []string{"EUW1_001"})
	if err == nil {
		t.Fatal("expected error from LLM")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error = %q, want to contain 'rate limited'", err.Error())
	}
}

func TestCoachStoreGetError(t *testing.T) {
	llm := &mockLLM{response: "advice"}
	st := &mockSessionStore{getErr: errors.New("connection refused")}
	svc := NewService(llm, st)

	a := makeTestAnalysis()
	_, err := svc.Coach(context.Background(), a, []string{"EUW1_001"})
	if err == nil {
		t.Fatal("expected error from store")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q, want to contain 'connection refused'", err.Error())
	}
}

func TestCoachStoreSaveError(t *testing.T) {
	llm := &mockLLM{response: "advice"}
	st := &mockSessionStore{saveErr: errors.New("disk full")}
	svc := NewService(llm, st)

	a := makeTestAnalysis()
	_, err := svc.Coach(context.Background(), a, []string{"EUW1_001"})
	if err == nil {
		t.Fatal("expected error from save")
	}
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("error = %q, want to contain 'disk full'", err.Error())
	}
}

func TestCoachSavedSessionAnalysisRoundTrips(t *testing.T) {
	llm := &mockLLM{response: "advice"}
	st := &mockSessionStore{}
	svc := NewService(llm, st)

	a := makeTestAnalysis()
	matchIDs := []string{"EUW1_001", "EUW1_002"}

	_, err := svc.Coach(context.Background(), a, matchIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var savedAnalysis map[string]interface{}
	if err := json.Unmarshal(st.savedSession.Analysis, &savedAnalysis); err != nil {
		t.Fatalf("saved analysis is not valid JSON: %v", err)
	}
	if savedAnalysis["puuid"] != "test-puuid" {
		t.Errorf("saved analysis puuid = %v, want %q", savedAnalysis["puuid"], "test-puuid")
	}

	var savedMatchIDs []string
	if err := json.Unmarshal(st.savedSession.MatchIDs, &savedMatchIDs); err != nil {
		t.Fatalf("saved match IDs is not valid JSON: %v", err)
	}
	if len(savedMatchIDs) != 2 || savedMatchIDs[0] != "EUW1_001" {
		t.Errorf("saved matchIDs = %v, want [EUW1_001, EUW1_002]", savedMatchIDs)
	}
}

func TestCoachEmptyMatchIDs(t *testing.T) {
	llm := &mockLLM{response: "advice"}
	st := &mockSessionStore{}
	svc := NewService(llm, st)

	a := makeTestAnalysis()

	resp, err := svc.Coach(context.Background(), a, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.NewMatches != 0 {
		t.Errorf("newMatches = %d, want 0", resp.NewMatches)
	}
	if st.savedSession.LatestMatchID != "" {
		t.Errorf("saved latestMatchID = %q, want empty", st.savedSession.LatestMatchID)
	}
}

func makeSessionFromAnalysis(t *testing.T, pa *analysis.PlayerAnalysis, matchIDs []string, createdAt time.Time) store.CoachingSession {
	t.Helper()
	analysisJSON, err := json.Marshal(pa)
	if err != nil {
		t.Fatalf("marshal analysis: %v", err)
	}
	matchIDsJSON, err := json.Marshal(matchIDs)
	if err != nil {
		t.Fatalf("marshal matchIDs: %v", err)
	}
	return store.CoachingSession{
		PUUID:         pa.PUUID,
		LatestMatchID: matchIDs[0],
		MatchIDs:      matchIDsJSON,
		Analysis:      analysisJSON,
		Advice:        "some advice",
		CreatedAt:     createdAt,
	}
}

func TestGetProgressMultipleSessions(t *testing.T) {
	session1Analysis := makeTestAnalysis()
	session1Analysis.Averages.KDA = 2.5
	session1Analysis.WinRate = 0.50
	session1Analysis.Tier = "GOLD"
	session1Analysis.Rank = "III"

	session2Analysis := makeTestAnalysis()
	session2Analysis.Averages.KDA = 3.5
	session2Analysis.WinRate = 0.60
	session2Analysis.Tier = "GOLD"
	session2Analysis.Rank = "II"

	now := time.Now()
	st := &mockSessionStore{
		sessions: []store.CoachingSession{
			makeSessionFromAnalysis(t, session1Analysis, []string{"EUW1_001", "EUW1_002"}, now.Add(-7*24*time.Hour)),
			makeSessionFromAnalysis(t, session2Analysis, []string{"EUW1_003", "EUW1_004", "EUW1_005"}, now),
		},
	}
	svc := NewService(nil, st)

	progress, err := svc.GetProgress(context.Background(), "test-puuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if progress.Sessions != 2 {
		t.Errorf("sessions = %d, want 2", progress.Sessions)
	}
	if progress.GameName != "TestPlayer" {
		t.Errorf("gameName = %q, want %q", progress.GameName, "TestPlayer")
	}
	if len(progress.Trend) != 2 {
		t.Fatalf("trend length = %d, want 2", len(progress.Trend))
	}

	if progress.Trend[0].MatchCount != 2 {
		t.Errorf("trend[0].matchCount = %d, want 2", progress.Trend[0].MatchCount)
	}
	if progress.Trend[0].WinRate != 0.50 {
		t.Errorf("trend[0].winRate = %f, want 0.50", progress.Trend[0].WinRate)
	}
	if progress.Trend[0].Tier != "GOLD" || progress.Trend[0].Rank != "III" {
		t.Errorf("trend[0].rank = %s %s, want GOLD III", progress.Trend[0].Tier, progress.Trend[0].Rank)
	}
	if progress.Trend[0].Averages.KDA != 2.5 {
		t.Errorf("trend[0].averages.kda = %f, want 2.5", progress.Trend[0].Averages.KDA)
	}

	if progress.Trend[1].MatchCount != 3 {
		t.Errorf("trend[1].matchCount = %d, want 3", progress.Trend[1].MatchCount)
	}
	if progress.Trend[1].Averages.KDA != 3.5 {
		t.Errorf("trend[1].averages.kda = %f, want 3.5", progress.Trend[1].Averages.KDA)
	}
}

func TestGetProgressNoSessions(t *testing.T) {
	st := &mockSessionStore{}
	svc := NewService(nil, st)

	progress, err := svc.GetProgress(context.Background(), "test-puuid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if progress.Sessions != 0 {
		t.Errorf("sessions = %d, want 0", progress.Sessions)
	}
	if len(progress.Trend) != 0 {
		t.Errorf("trend length = %d, want 0", len(progress.Trend))
	}
}

func TestGetProgressNilStore(t *testing.T) {
	svc := NewService(nil, nil)

	_, err := svc.GetProgress(context.Background(), "test-puuid")
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestGetProgressStoreError(t *testing.T) {
	st := &mockSessionStore{sessionsErr: errors.New("connection lost")}
	svc := NewService(nil, st)

	_, err := svc.GetProgress(context.Background(), "test-puuid")
	if err == nil {
		t.Fatal("expected error from store")
	}
	if !strings.Contains(err.Error(), "connection lost") {
		t.Errorf("error = %q, want to contain 'connection lost'", err.Error())
	}
}
