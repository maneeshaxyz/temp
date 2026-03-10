package intake

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	outboundStatusSuccess = "success"
	outboundAPIKeyEnvVar  = "X_API_KEY"
	outboundAPIKeyHeader  = "X-API-Key"
	outboundURLFormat     = "http://%s:8888/api/results"
	outboundHTTPTimeout   = 5 * time.Second
)

type Handler struct {
	client *http.Client
	apiKey string
}

func NewHandler() *Handler {
	return newHandlerWithDeps(
		&http.Client{Timeout: outboundHTTPTimeout},
		strings.TrimSpace(os.Getenv(outboundAPIKeyEnvVar)),
	)
}

func newHandlerWithDeps(client *http.Client, apiKey string) *Handler {
	if client == nil {
		client = &http.Client{Timeout: outboundHTTPTimeout}
	}
	return &Handler{
		client: client,
		apiKey: strings.TrimSpace(apiKey),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeEventsResponse(w, http.StatusMethodNotAllowed, errMethodNotAllowed)
		return
	}

	if !isJSONContentType(r.Header.Get("Content-Type")) {
		writeEventsResponse(w, http.StatusUnsupportedMediaType, errContentTypeJSON)
		return
	}

	body, err := decodeJSONBody(r.Body)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			writeEventsResponse(w, http.StatusBadRequest, errBodyRequired)
		case errors.Is(err, errMultipleJSONObjects):
			writeEventsResponse(w, http.StatusBadRequest, errSingleJSONObject)
		case errors.Is(err, errMissingTimestamp):
			writeEventsResponse(w, http.StatusBadRequest, errTimestampRequired)
		case errors.Is(err, errTimestampNotString):
			writeEventsResponse(w, http.StatusBadRequest, errTimestampString)
		case errors.Is(err, errTimestampNotRFC3339):
			writeEventsResponse(w, http.StatusBadRequest, errTimestampRFC3339)
		case errors.Is(err, errMissingInstanceID):
			writeEventsResponse(w, http.StatusBadRequest, errInstanceIDRequired)
		case errors.Is(err, errInstanceIDNotString):
			writeEventsResponse(w, http.StatusBadRequest, errInstanceIDString)
		case errors.Is(err, errInstanceIDEmptyErr):
			writeEventsResponse(w, http.StatusBadRequest, errInstanceIDEmpty)
		case errors.Is(err, errMissingSigVersion):
			writeEventsResponse(w, http.StatusBadRequest, errSigVersionRequired)
		case errors.Is(err, errSigVersionNotString):
			writeEventsResponse(w, http.StatusBadRequest, errSigVersionString)
		case errors.Is(err, errSigVersionEmptyErr):
			writeEventsResponse(w, http.StatusBadRequest, errSigVersionEmpty)
		case errors.Is(err, errMissingSigUpdatedAt):
			writeEventsResponse(w, http.StatusBadRequest, errSigUpdatedAtRequired)
		case errors.Is(err, errSigUpdatedAtNotStr):
			writeEventsResponse(w, http.StatusBadRequest, errSigUpdatedAtString)
		case errors.Is(err, errSigUpdatedAtRFC3339Err):
			writeEventsResponse(w, http.StatusBadRequest, errSigUpdatedAtRFC3339)
		default:
			writeEventsResponse(w, http.StatusBadRequest, errInvalidJSONText)
		}
		return
	}

	go h.sendResults(body)
	writeEventsResponse(w, http.StatusAccepted, "")
}

func writeEventsResponse(w http.ResponseWriter, status int, message string) {
	if status == http.StatusAccepted {
		writeStatusAccepted(w)
		return
	}

	writeError(w, status, message)
}

type outboundResultsPayload struct {
	Status    string         `json:"status"`
	Data      map[string]any `json:"data"`
	Timestamp string         `json:"timestamp"`
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

	resp, err := h.client.Do(req)
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
