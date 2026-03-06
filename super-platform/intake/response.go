package intake

import (
	"encoding/json"
	"net/http"
)

const (
	errMethodNotAllowed  = "method not allowed"
	errContentTypeJSON   = "content type must be application/json"
	errBodyRequired      = "request body is required"
	errInvalidJSONText   = "invalid JSON"
	errSingleJSONObject  = "request body must contain only one JSON object"
	errTimestampRequired = "timestamp is required"
	errTimestampString   = "timestamp must be a string"
	errTimestampRFC3339  = "timestamp must be RFC3339"
)

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeStatusAccepted(w http.ResponseWriter) {
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
