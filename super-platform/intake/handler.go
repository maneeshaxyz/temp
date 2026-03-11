package intake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	EventsPath            = "/v1/silver/events"
	outboundStatusSuccess = "success"
	outboundAPIKeyEnvVar  = "X_API_KEY"
	outboundAPIKeyHeader  = "X-API-Key"
	outboundURLFormat     = "http://%s:8888/api/results"
	outboundHTTPTimeout   = 5 * time.Second

	errMethodNotAllowed     = "method not allowed"
	errContentTypeJSON      = "content type must be application/json"
	errBodyRequired         = "request body is required"
	errInvalidJSONText      = "invalid JSON"
	errTimestampRequired    = "timestamp is required"
	errTimestampRFC3339     = "timestamp must be RFC3339"
	errInstanceIDRequired   = "instance_id is required"
	errSigVersionRequired   = "signature_version is required"
	errSigUpdatedAtRequired = "signature_updated_at is required"
	errSigUpdatedAtRFC3339  = "signature_updated_at must be RFC3339"
)

type Handler struct {
	client *http.Client
	apiKey string
}

type inboundEvent struct {
	Timestamp          string `json:"timestamp"`
	InstanceID         string `json:"instance_id"`
	SignatureVersion   string `json:"signature_version"`
	SignatureUpdatedAt string `json:"signature_updated_at"`
}

type outboundResultsPayload struct {
	Status    string         `json:"status"`
	Data      map[string]any `json:"data"`
	Timestamp string         `json:"timestamp"`
}

func NewHandler() *Handler {
	return &Handler{
		client: &http.Client{Timeout: outboundHTTPTimeout},
		apiKey: strings.TrimSpace(os.Getenv(outboundAPIKeyEnvVar)),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, errMethodNotAllowed)
		return
	}

	if !isJSONContentType(r.Header.Get("Content-Type")) {
		writeError(w, http.StatusUnsupportedMediaType, errContentTypeJSON)
		return
	}

	event, errMessage := decodeAndValidateEvent(r.Body)
	if errMessage != "" {
		writeError(w, http.StatusBadRequest, errMessage)
		return
	}

	go h.sendResults(event)
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (h *Handler) sendResults(event inboundEvent) {
	if h.apiKey == "" {
		slog.Error("skipping outbound results callback: missing API key", "env", outboundAPIKeyEnvVar)
		return
	}

	payload := outboundResultsPayload{
		Status: outboundStatusSuccess,
		Data: map[string]any{
			"timestamp":            event.Timestamp,
			"instance_id":          event.InstanceID,
			"signature_version":    event.SignatureVersion,
			"signature_updated_at": event.SignatureUpdatedAt,
		},
		Timestamp: event.Timestamp,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal outbound results payload", "error", err, "instance_id", event.InstanceID)
		return
	}

	destination := buildResultsURL(event.InstanceID)
	req, err := http.NewRequest(http.MethodPost, destination, bytes.NewReader(body))
	if err != nil {
		slog.Error("failed to build outbound results request", "error", err, "destination", destination, "instance_id", event.InstanceID)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(outboundAPIKeyHeader, h.apiKey)

	client := h.client
	if client == nil {
		client = &http.Client{Timeout: outboundHTTPTimeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("outbound results callback failed", "error", err, "destination", destination, "instance_id", event.InstanceID)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		errorBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if err != nil {
			slog.Error(
				"outbound results callback returned non-2xx and failed to read body",
				"status_code", resp.StatusCode,
				"destination", destination,
				"instance_id", event.InstanceID,
				"read_error", err,
			)
			return
		}
		slog.Error(
			"outbound results callback returned non-2xx",
			"status_code", resp.StatusCode,
			"destination", destination,
			"instance_id", event.InstanceID,
			"body", strings.TrimSpace(string(errorBody)),
		)
	}
}

func buildResultsURL(instanceID string) string {
	return fmt.Sprintf(outboundURLFormat, instanceID)
}

func decodeAndValidateEvent(r io.Reader) (inboundEvent, string) {
	var event inboundEvent

	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&event); err != nil {
		if err == io.EOF {
			return inboundEvent{}, errBodyRequired
		}
		return inboundEvent{}, errInvalidJSONText
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return inboundEvent{}, errInvalidJSONText
	}

	event.Timestamp = strings.TrimSpace(event.Timestamp)
	event.InstanceID = strings.TrimSpace(event.InstanceID)
	event.SignatureVersion = strings.TrimSpace(event.SignatureVersion)
	event.SignatureUpdatedAt = strings.TrimSpace(event.SignatureUpdatedAt)

	if event.Timestamp == "" {
		return inboundEvent{}, errTimestampRequired
	}
	if _, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		return inboundEvent{}, errTimestampRFC3339
	}
	if event.InstanceID == "" {
		return inboundEvent{}, errInstanceIDRequired
	}
	if event.SignatureVersion == "" {
		return inboundEvent{}, errSigVersionRequired
	}
	if event.SignatureUpdatedAt == "" {
		return inboundEvent{}, errSigUpdatedAtRequired
	}
	if _, err := time.Parse(time.RFC3339, event.SignatureUpdatedAt); err != nil {
		return inboundEvent{}, errSigUpdatedAtRFC3339
	}

	return event, ""
}

func isJSONContentType(contentType string) bool {
	if contentType == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	return mediaType == "application/json"
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
