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
			name:        "missing required field",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        `{"timestamp":"2026-03-05T10:30:45Z","signature_version":"daily.cld:0","signature_updated_at":"2026-03-08T07:57:37Z"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"instance_id is required"}`,
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
			h := &Handler{apiKey: ""}

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
