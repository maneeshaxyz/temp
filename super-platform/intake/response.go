package intake

import (
	"encoding/json"
	"net/http"
)

const (
	errMethodNotAllowed     = "method not allowed"
	errContentTypeJSON      = "content type must be application/json"
	errBodyRequired         = "request body is required"
	errInvalidJSONText      = "invalid JSON"
	errSingleJSONObject     = "request body must contain only one JSON object"
	errTimestampRequired    = "timestamp is required"
	errTimestampString      = "timestamp must be a string"
	errTimestampRFC3339     = "timestamp must be RFC3339"
	errInstanceIDRequired   = "instance_id is required"
	errInstanceIDString     = "instance_id must be a string"
	errInstanceIDEmpty      = "instance_id must not be empty"
	errSigVersionRequired   = "signature_version is required"
	errSigVersionString     = "signature_version must be a string"
	errSigVersionEmpty      = "signature_version must not be empty"
	errSigUpdatedAtRequired = "signature_updated_at is required"
	errSigUpdatedAtString   = "signature_updated_at must be a string"
	errSigUpdatedAtRFC3339  = "signature_updated_at must be RFC3339"
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
