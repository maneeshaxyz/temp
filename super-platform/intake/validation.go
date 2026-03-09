package intake

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"time"
)

var (
	errInvalidJSON            = errors.New("invalid JSON")
	errMultipleJSONObjects    = errors.New("request body must contain only one JSON object")
	errMissingTimestamp       = errors.New("timestamp is required")
	errTimestampNotString     = errors.New("timestamp must be a string")
	errTimestampNotRFC3339    = errors.New("timestamp must be RFC3339")
	errMissingInstanceID      = errors.New("instance_id is required")
	errInstanceIDNotString    = errors.New("instance_id must be a string")
	errInstanceIDEmptyErr     = errors.New("instance_id must not be empty")
	errMissingSigVersion      = errors.New("signature_version is required")
	errSigVersionNotString    = errors.New("signature_version must be a string")
	errSigVersionEmptyErr     = errors.New("signature_version must not be empty")
	errMissingSigUpdatedAt    = errors.New("signature_updated_at is required")
	errSigUpdatedAtNotStr     = errors.New("signature_updated_at must be a string")
	errSigUpdatedAtRFC3339Err = errors.New("signature_updated_at must be RFC3339")
)

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

func decodeJSONBody(r io.Reader) (map[string]any, error) {
	var body map[string]any
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, errInvalidJSON
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, errMultipleJSONObjects
	}

	tsRaw, ok := body["timestamp"]
	if !ok {
		return nil, errMissingTimestamp
	}
	ts, ok := tsRaw.(string)
	if !ok {
		return nil, errTimestampNotString
	}

	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		return nil, errTimestampNotRFC3339
	}

	instanceIDRaw, ok := body["instance_id"]
	if !ok {
		return nil, errMissingInstanceID
	}
	instanceID, ok := instanceIDRaw.(string)
	if !ok {
		return nil, errInstanceIDNotString
	}
	if instanceID == "" {
		return nil, errInstanceIDEmptyErr
	}

	sigVersionRaw, ok := body["signature_version"]
	if !ok {
		return nil, errMissingSigVersion
	}
	sigVersion, ok := sigVersionRaw.(string)
	if !ok {
		return nil, errSigVersionNotString
	}
	if sigVersion == "" {
		return nil, errSigVersionEmptyErr
	}

	sigUpdatedAtRaw, ok := body["signature_updated_at"]
	if !ok {
		return nil, errMissingSigUpdatedAt
	}
	sigUpdatedAt, ok := sigUpdatedAtRaw.(string)
	if !ok {
		return nil, errSigUpdatedAtNotStr
	}
	if _, err := time.Parse(time.RFC3339, sigUpdatedAt); err != nil {
		return nil, errSigUpdatedAtRFC3339Err
	}

	return body, nil
}
