package intake

import (
	"errors"
	"io"
	"net/http"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
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

	if _, err := decodeJSONBody(r.Body); err != nil {
		switch {
		case errors.Is(err, io.EOF):
			writeError(w, http.StatusBadRequest, errBodyRequired)
		case errors.Is(err, errMultipleJSONObjects):
			writeError(w, http.StatusBadRequest, errSingleJSONObject)
		case errors.Is(err, errMissingTimestamp):
			writeError(w, http.StatusBadRequest, errTimestampRequired)
		case errors.Is(err, errTimestampNotString):
			writeError(w, http.StatusBadRequest, errTimestampString)
		case errors.Is(err, errTimestampNotRFC3339):
			writeError(w, http.StatusBadRequest, errTimestampRFC3339)
		default:
			writeError(w, http.StatusBadRequest, errInvalidJSONText)
		}
		return
	}

	writeStatusAccepted(w)
}
