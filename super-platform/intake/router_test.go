package intake

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewMuxRoutesEventsPath(t *testing.T) {
	mux := NewMux()

	req := httptest.NewRequest(http.MethodPost, EventsPath, strings.NewReader(`{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if strings.TrimSpace(rr.Body.String()) != `{"ok":true}` {
		t.Fatalf("body = %q, want %q", strings.TrimSpace(rr.Body.String()), `{"ok":true}`)
	}
}

func TestNewMuxUnknownPath(t *testing.T) {
	mux := NewMux()

	req := httptest.NewRequest(http.MethodPost, "/unknown", strings.NewReader(`{"timestamp":"2026-03-05T10:30:45Z"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
