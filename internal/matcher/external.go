package matcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// ResponseType defines the type of response parsing.
type ResponseType string

const (
	// ResponseTypeJSON expects JSON response with flagged/hits structure.
	ResponseTypeJSON ResponseType = "json"
	// ResponseTypeSimple treats any non-empty response as flagged.
	ResponseTypeSimple ResponseType = "simple"
)

// ExternalConfig is the configuration for the external API matcher.
type ExternalConfig struct {
	Endpoint     string        `mapstructure:"endpoint"`
	APIKey       string        `mapstructure:"api_key"`
	Timeout      time.Duration `mapstructure:"timeout"`
	ResponseType ResponseType  `mapstructure:"response_type"`
}

// ExternalMatcher implements the Matcher interface by delegating to an external API.
type ExternalMatcher struct {
	cfg    *ExternalConfig
	client *http.Client
}

// NewExternal creates a new external API matcher.
func NewExternal(cfg *ExternalConfig) (Matcher, error) {
	if cfg == nil {
		return nil, fmt.Errorf("external config is required")
	}
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("external endpoint is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}

	return &ExternalMatcher{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}, nil
}

// Name returns the matcher name.
func (m *ExternalMatcher) Name() string {
	return "external-adapter"
}

// Match sends the text to the external API and returns the parsed hits.
func (m *ExternalMatcher) Match(text string, wordBank storage.WordBank) ([]model.HitRecord, error) {
	body, err := m.callAPI(text)
	if err != nil {
		// Fail open: return nil hits on error
		return nil, nil
	}

	return m.parseResponse(body)
}

// callAPI makes the HTTP request to the external API.
func (m *ExternalMatcher) callAPI(text string) ([]byte, error) {
	reqBody, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, m.cfg.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if m.cfg.APIKey != "" {
		req.Header.Set("Authorization", m.cfg.APIKey)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

// parseResponse parses the API response based on the configured response type.
func (m *ExternalMatcher) parseResponse(body []byte) ([]model.HitRecord, error) {
	switch m.cfg.ResponseType {
	case ResponseTypeJSON:
		return m.parseJSONResponse(body)
	case ResponseTypeSimple:
		return m.parseSimpleResponse(body)
	default:
		return m.parseJSONResponse(body)
	}
}

// JSONResponse represents the expected JSON response format.
type JSONResponse struct {
	Flagged bool      `json:"flagged"`
	Hits    []JSONHit `json:"hits"`
}

// JSONHit represents a hit in the JSON response.
type JSONHit struct {
	Word     string `json:"word"`
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

// parseJSONResponse parses a JSON response.
func (m *ExternalMatcher) parseJSONResponse(body []byte) ([]model.HitRecord, error) {
	var resp JSONResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if !resp.Flagged {
		return nil, nil
	}

	var hits []model.HitRecord
	for _, h := range resp.Hits {
		severity := model.Severity(strings.ToLower(h.Severity))
		hits = append(hits, model.HitRecord{
			Word:     h.Word,
			Type:     h.Type,
			Severity: severity,
		})
	}

	return hits, nil
}

// parseSimpleResponse treats any non-empty response as flagged.
func (m *ExternalMatcher) parseSimpleResponse(body []byte) ([]model.HitRecord, error) {
	if len(body) == 0 {
		return nil, nil
	}
	// For simple mode, we can't determine specific hit details
	return []model.HitRecord{{Word: "", Type: "", Severity: model.SeverityMedium}}, nil
}
