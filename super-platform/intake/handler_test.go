package intake

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerResponses(t *testing.T) {
	validPayload := `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`

	tests := []struct {
		name          string
		method        string
		contentType   string
		body          string
		wantStatus    int
		wantBody      string
		wantAllowPost bool
	}{
		{
			name:          "method not allowed",
			method:        http.MethodGet,
			contentType:   "application/json",
			body:          validPayload,
			wantStatus:    http.StatusMethodNotAllowed,
			wantBody:      `{"error":"method not allowed"}`,
			wantAllowPost: true,
		},
		{
			name:        "unsupported media type",
			method:      http.MethodPost,
			contentType: "text/plain",
			body:        validPayload,
			wantStatus:  http.StatusUnsupportedMediaType,
			wantBody:    `{"error":"content type must be application/json"}`,
		},
		{
			name:        "empty body",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "",
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"request body is required"}`,
		},
		{
			name:        "invalid json",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"invalid JSON"}`,
		},
		{
			name:        "multiple json objects",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        validPayload + `{"another":1}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"request body must contain only one JSON object"}`,
		},
		{
			name:        "timestamp missing",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"event":"scan_complete"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"timestamp is required"}`,
		},
		{
			name:        "timestamp not string",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":123,"instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"timestamp must be a string"}`,
		},
		{
			name:        "timestamp not rfc3339",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026/03/05 10:30:45","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"timestamp must be RFC3339"}`,
		},
		{
			name:        "instance_id missing",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"instance_id is required"}`,
		},
		{
			name:        "instance_id not string",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":123,"signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"instance_id must be a string"}`,
		},
		{
			name:        "instance_id empty",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"instance_id must not be empty"}`,
		},
		{
			name:        "signature_version missing",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_version is required"}`,
		},
		{
			name:        "signature_version not string",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":123,"signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_version must be a string"}`,
		},
		{
			name:        "signature_version empty",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_version must not be empty"}`,
		},
		{
			name:        "signature_updated_at missing",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_updated_at is required"}`,
		},
		{
			name:        "signature_updated_at not string",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":123}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_updated_at must be a string"}`,
		},
		{
			name:        "signature_updated_at not rfc3339",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","instance_id":"172.25.0.19","signature_version":"daily.cld:0","signature_updated_at":"2026/03/08 07:57:37"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"signature_updated_at must be RFC3339"}`,
		},
		{
			name:        "accepted",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        validPayload,
			wantStatus:  http.StatusAccepted,
			wantBody:    `{"ok":true}`,
		},
		{
			name:        "content type with charset",
			method:      http.MethodPost,
			contentType: "application/json; charset=utf-8",
			body:        validPayload,
			wantStatus:  http.StatusAccepted,
			wantBody:    `{"ok":true}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler()

			req := httptest.NewRequest(tc.method, EventsPath, strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tc.wantStatus)
			}
			if strings.TrimSpace(rr.Body.String()) != tc.wantBody {
				t.Fatalf("body = %q, want %q", strings.TrimSpace(rr.Body.String()), tc.wantBody)
			}
			if tc.wantAllowPost && rr.Header().Get("Allow") != http.MethodPost {
				t.Fatalf("Allow header = %q, want %q", rr.Header().Get("Allow"), http.MethodPost)
			}
		})
	}
}
