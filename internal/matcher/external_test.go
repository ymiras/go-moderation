package matcher

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ymiras/go-moderation/internal/model"
	"github.com/ymiras/go-moderation/internal/storage"
)

// mockWordBankForExternal implements storage.WordBank for external testing.
type mockWordBankForExternal struct{}

func (m *mockWordBankForExternal) Load(path string) error                      { return nil }
func (m *mockWordBankForExternal) Contains(word string) (bool, *model.Keyword) { return false, nil }
func (m *mockWordBankForExternal) Words() []*model.Keyword                     { return nil }
func (m *mockWordBankForExternal) Size() int                                   { return 0 }

var _ storage.WordBank = (*mockWordBankForExternal)(nil)

func TestExternalMatcher_Match(t *testing.T) {
	t.Run("successful json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := JSONResponse{
				Flagged: true,
				Hits: []JSONHit{
					{Word: "bad", Type: "profanity", Severity: "high"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := &ExternalConfig{
			Endpoint:     server.URL,
			Timeout:      5 * time.Second,
			ResponseType: ResponseTypeJSON,
		}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}

		hits, err := m.Match("hello bad world", &mockWordBankForExternal{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 1 {
			t.Fatalf("expected 1 hit, got %d", len(hits))
		}
		if hits[0].Word != "bad" {
			t.Errorf("expected word 'bad', got '%s'", hits[0].Word)
		}
	})

	t.Run("not flagged response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := JSONResponse{Flagged: false, Hits: nil}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := &ExternalConfig{
			Endpoint:     server.URL,
			Timeout:      5 * time.Second,
			ResponseType: ResponseTypeJSON,
		}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}

		hits, err := m.Match("hello world", &mockWordBankForExternal{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 0 {
			t.Errorf("expected 0 hits, got %d", len(hits))
		}
	})

	t.Run("simple response type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("flagged"))
		}))
		defer server.Close()

		cfg := &ExternalConfig{
			Endpoint:     server.URL,
			Timeout:      5 * time.Second,
			ResponseType: ResponseTypeSimple,
		}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}

		hits, err := m.Match("hello world", &mockWordBankForExternal{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(hits) != 1 {
			t.Fatalf("expected 1 hit for simple response, got %d", len(hits))
		}
	})

	t.Run("fail open on error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := &ExternalConfig{
			Endpoint:     server.URL,
			Timeout:      5 * time.Second,
			ResponseType: ResponseTypeJSON,
		}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}

		hits, err := m.Match("hello world", &mockWordBankForExternal{})
		// Should fail open - return nil hits, nil error
		if err != nil {
			t.Fatalf("expected fail open (nil error), got error: %v", err)
		}
		if hits != nil {
			t.Errorf("expected nil hits on fail open, got %v", hits)
		}
	})

	t.Run("fail open on timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Write([]byte("ok"))
		}))
		defer server.Close()

		cfg := &ExternalConfig{
			Endpoint:     server.URL,
			Timeout:      10 * time.Millisecond, // Very short timeout
			ResponseType: ResponseTypeJSON,
		}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}

		hits, err := m.Match("hello world", &mockWordBankForExternal{})
		// Should fail open on timeout
		if err != nil {
			t.Fatalf("expected fail open (nil error), got error: %v", err)
		}
		if hits != nil {
			t.Errorf("expected nil hits on timeout, got %v", hits)
		}
	})
}

func TestExternalMatcher_Name(t *testing.T) {
	cfg := &ExternalConfig{
		Endpoint:     "http://example.com",
		Timeout:      5 * time.Second,
		ResponseType: ResponseTypeJSON,
	}
	m, err := NewExternal(cfg)
	if err != nil {
		t.Fatalf("failed to create matcher: %v", err)
	}

	if m.Name() != "external-adapter" {
		t.Errorf("expected 'external-adapter', got '%s'", m.Name())
	}
}

func TestExternalMatcher_NewExternal(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		_, err := NewExternal(nil)
		if err == nil {
			t.Error("expected error for nil config")
		}
	})

	t.Run("empty endpoint", func(t *testing.T) {
		_, err := NewExternal(&ExternalConfig{Endpoint: ""})
		if err == nil {
			t.Error("expected error for empty endpoint")
		}
	})

	t.Run("default timeout", func(t *testing.T) {
		cfg := &ExternalConfig{Endpoint: "http://example.com"}
		m, err := NewExternal(cfg)
		if err != nil {
			t.Fatalf("failed to create matcher: %v", err)
		}
		if m.(*ExternalMatcher).cfg.Timeout != 5*time.Second {
			t.Errorf("expected default timeout 5s, got %v", m.(*ExternalMatcher).cfg.Timeout)
		}
	})
}
