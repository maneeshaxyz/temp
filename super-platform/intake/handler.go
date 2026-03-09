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
		case errors.Is(err, errMissingInstanceID):
			writeError(w, http.StatusBadRequest, errInstanceIDRequired)
		case errors.Is(err, errInstanceIDNotString):
			writeError(w, http.StatusBadRequest, errInstanceIDString)
		case errors.Is(err, errInstanceIDEmptyErr):
			writeError(w, http.StatusBadRequest, errInstanceIDEmpty)
		case errors.Is(err, errMissingSigVersion):
			writeError(w, http.StatusBadRequest, errSigVersionRequired)
		case errors.Is(err, errSigVersionNotString):
			writeError(w, http.StatusBadRequest, errSigVersionString)
		case errors.Is(err, errSigVersionEmptyErr):
			writeError(w, http.StatusBadRequest, errSigVersionEmpty)
		case errors.Is(err, errMissingSigUpdatedAt):
			writeError(w, http.StatusBadRequest, errSigUpdatedAtRequired)
		case errors.Is(err, errSigUpdatedAtNotStr):
			writeError(w, http.StatusBadRequest, errSigUpdatedAtString)
		case errors.Is(err, errSigUpdatedAtRFC3339Err):
			writeError(w, http.StatusBadRequest, errSigUpdatedAtRFC3339)
		default:
			writeError(w, http.StatusBadRequest, errInvalidJSONText)
		}
		return
	}

	writeStatusAccepted(w)
}
