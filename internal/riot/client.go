package riot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/HatiCode/league-buddy/internal/models"
	"github.com/HatiCode/league-buddy/pkg/ratelimit"
)

// ClientOption configures the client.
type ClientOption func(*APIClient)

// WithBaseURL overrides the base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *APIClient) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *APIClient) {
		c.httpClient = client
	}
}

// WithRateLimiter adds rate limiting to the client.
func WithRateLimiter(limiter *ratelimit.Limiter) ClientOption {
	return func(c *APIClient) {
		c.httpClient.Transport = ratelimit.NewRoundTripper(limiter, c.httpClient.Transport)
	}
}

// APIClient implements the Client interface.
type APIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Riot API client.
func NewClient(apiKey string, opts ...ClientOption) *APIClient {
	c := &APIClient{
		apiKey:  apiKey,
		baseURL: "",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// GetSummonerByName fetches a summoner by their name.
func (c *APIClient) GetSummonerByName(ctx context.Context, region, name string) (*models.Summoner, error) {
	if !isValidPlatform(region) {
		return nil, ErrInvalidRegion
	}

	path := fmt.Sprintf("/lol/summoner/v4/summoners/by-name/%s", url.PathEscape(name))
	var summoner models.Summoner
	if err := c.get(ctx, region, path, &summoner); err != nil {
		return nil, err
	}
	return &summoner, nil
}

// GetSummonerByPUUID fetches a summoner by their PUUID.
func (c *APIClient) GetSummonerByPUUID(ctx context.Context, region, puuid string) (*models.Summoner, error) {
	if !isValidPlatform(region) {
		return nil, ErrInvalidRegion
	}

	path := fmt.Sprintf("/lol/summoner/v4/summoners/by-puuid/%s", puuid)
	var summoner models.Summoner
	if err := c.get(ctx, region, path, &summoner); err != nil {
		return nil, err
	}
	return &summoner, nil
}

// GetMatchIDs fetches recent match IDs for a player.
func (c *APIClient) GetMatchIDs(ctx context.Context, platform, puuid string, count int) ([]string, error) {
	if !isValidPlatform(platform) {
		return nil, ErrInvalidRegion
	}

	region := PlatformToRegion[platform]
	path := fmt.Sprintf("/lol/match/v5/matches/by-puuid/%s/ids?count=%s", puuid, strconv.Itoa(count))

	var matchIDs []string
	if err := c.getRegional(ctx, region, path, &matchIDs); err != nil {
		return nil, err
	}
	return matchIDs, nil
}

// GetMatch fetches full match details.
func (c *APIClient) GetMatch(ctx context.Context, platform, matchID string) (*models.Match, error) {
	if !isValidPlatform(platform) {
		return nil, ErrInvalidRegion
	}

	region := PlatformToRegion[platform]
	path := fmt.Sprintf("/lol/match/v5/matches/%s", matchID)

	var match models.Match
	if err := c.getRegional(ctx, region, path, &match); err != nil {
		return nil, err
	}
	return &match, nil
}

// GetLeagueEntries fetches ranked entries for a summoner.
func (c *APIClient) GetLeagueEntries(ctx context.Context, region, summonerID string) ([]models.LeagueEntry, error) {
	if !isValidPlatform(region) {
		return nil, ErrInvalidRegion
	}

	path := fmt.Sprintf("/lol/league/v4/entries/by-summoner/%s", summonerID)
	var entries []models.LeagueEntry
	if err := c.get(ctx, region, path, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// get performs a GET request to platform-specific endpoints.
func (c *APIClient) get(ctx context.Context, platform, path string, result any) error {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.api.riotgames.com", platform)
	}
	return c.doRequest(ctx, baseURL+path, result)
}

// getRegional performs a GET request to regional endpoints (for match-v5).
func (c *APIClient) getRegional(ctx context.Context, region, path string, result any) error {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.api.riotgames.com", region)
	}
	return c.doRequest(ctx, baseURL+path, result)
}

// doRequest executes the HTTP request and handles common responses.
func (c *APIClient) doRequest(ctx context.Context, url string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-Riot-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return json.NewDecoder(resp.Body).Decode(result)
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusTooManyRequests:
		return ErrRateLimited
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// isValidPlatform checks if the platform is supported.
func isValidPlatform(platform string) bool {
	_, ok := PlatformToRegion[platform]
	return ok
}
