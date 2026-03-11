package intake

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHandlerSendsOutboundResultsRequest(t *testing.T) {
	type outboundRequest struct {
		Method      string
		URL         string
		ContentType string
		APIKey      string
		Body        []byte
	}

	reqCh := make(chan outboundRequest, 1)
	client := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			_ = req.Body.Close()
			reqCh <- outboundRequest{
				Method:      req.Method,
				URL:         req.URL.String(),
				ContentType: req.Header.Get("Content-Type"),
				APIKey:      req.Header.Get("X-API-Key"),
				Body:        body,
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	h := &Handler{client: client, apiKey: "test-api-key"}
	inbound := `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`

	req := httptest.NewRequest(http.MethodPost, EventsPath, strings.NewReader(inbound))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	var outbound outboundRequest
	select {
	case outbound = <-reqCh:
	case <-time.After(2 * time.Second):
		t.Errorf("timed out waiting for outbound callback")
		return
	}

	if outbound.Method != http.MethodPost {
		t.Errorf("method = %q, want %q", outbound.Method, http.MethodPost)
	}
	if outbound.URL != "http://172.25.0.19:8888/api/results" {
		t.Errorf("url = %q, want %q", outbound.URL, "http://172.25.0.19:8888/api/results")
	}
	if outbound.ContentType != "application/json" {
		t.Errorf("content type = %q, want %q", outbound.ContentType, "application/json")
	}
	if outbound.APIKey != "test-api-key" {
		t.Errorf("X-API-Key = %q, want %q", outbound.APIKey, "test-api-key")
	}

	var payload map[string]any
	if err := json.Unmarshal(outbound.Body, &payload); err != nil {
		t.Errorf("failed to unmarshal outbound body: %v", err)
		return
	}
	if payload["status"] != "success" {
		t.Errorf("status = %v, want %q", payload["status"], "success")
	}
	if payload["timestamp"] != "2026-03-05T10:30:45Z" {
		t.Errorf("timestamp = %v, want %q", payload["timestamp"], "2026-03-05T10:30:45Z")
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Errorf("data = %T, want map[string]any", payload["data"])
		return
	}
	if data["instance_id"] != "172.25.0.19" {
		t.Errorf("data.instance_id = %v, want %q", data["instance_id"], "172.25.0.19")
	}
	if data["signature_version"] != "daily.cld:0" {
		t.Errorf("data.signature_version = %v, want %q", data["signature_version"], "daily.cld:0")
	}
	if data["signature_updated_at"] != "2026-03-08T07:57:37Z" {
		t.Errorf("data.signature_updated_at = %v, want %q", data["signature_updated_at"], "2026-03-08T07:57:37Z")
	}
	if data["timestamp"] != "2026-03-05T10:30:45Z" {
		t.Errorf("data.timestamp = %v, want %q", data["timestamp"], "2026-03-05T10:30:45Z")
	}
}

func TestHandlerSkipsOutboundResultsWhenAPIKeyMissing(t *testing.T) {
	calledCh := make(chan struct{}, 1)
	client := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calledCh <- struct{}{}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	h := &Handler{client: client, apiKey: ""}
	inbound := `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`

	req := httptest.NewRequest(http.MethodPost, EventsPath, strings.NewReader(inbound))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}

	select {
	case <-calledCh:
		t.Fatal("outbound callback was sent even though API key is missing")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestBuildResultsURL(t *testing.T) {
	got := buildResultsURL("10.0.0.8")
	want := "http://10.0.0.8:8888/api/results"
	if got != want {
		t.Fatalf("buildResultsURL() = %q, want %q", got, want)
	}
}
