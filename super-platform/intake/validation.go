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

type inboundEvent struct {
	Timestamp          string
	InstanceID         string
	SignatureVersion   string
	SignatureUpdatedAt string
}

func decodeJSONBody(r io.Reader) (inboundEvent, error) {
	var body map[string]any
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return inboundEvent{}, io.EOF
		}
		return inboundEvent{}, errInvalidJSON
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return inboundEvent{}, errMultipleJSONObjects
	}

	tsRaw, ok := body["timestamp"]
	if !ok {
		return inboundEvent{}, errMissingTimestamp
	}
	ts, ok := tsRaw.(string)
	if !ok {
		return inboundEvent{}, errTimestampNotString
	}

	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		return inboundEvent{}, errTimestampNotRFC3339
	}

	instanceIDRaw, ok := body["instance_id"]
	if !ok {
		return inboundEvent{}, errMissingInstanceID
	}
	instanceID, ok := instanceIDRaw.(string)
	if !ok {
		return inboundEvent{}, errInstanceIDNotString
	}
	if instanceID == "" {
		return inboundEvent{}, errInstanceIDEmptyErr
	}

	sigVersionRaw, ok := body["signature_version"]
	if !ok {
		return inboundEvent{}, errMissingSigVersion
	}
	sigVersion, ok := sigVersionRaw.(string)
	if !ok {
		return inboundEvent{}, errSigVersionNotString
	}
	if sigVersion == "" {
		return inboundEvent{}, errSigVersionEmptyErr
	}

	sigUpdatedAtRaw, ok := body["signature_updated_at"]
	if !ok {
		return inboundEvent{}, errMissingSigUpdatedAt
	}
	sigUpdatedAt, ok := sigUpdatedAtRaw.(string)
	if !ok {
		return inboundEvent{}, errSigUpdatedAtNotStr
	}
	if _, err := time.Parse(time.RFC3339, sigUpdatedAt); err != nil {
		return inboundEvent{}, errSigUpdatedAtRFC3339Err
	}

	return inboundEvent{
		Timestamp:          ts,
		InstanceID:         instanceID,
		SignatureVersion:   sigVersion,
		SignatureUpdatedAt: sigUpdatedAt,
	}, nil
}
