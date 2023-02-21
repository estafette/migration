package migration

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const StatusUnset Status = -1

const (
	StatusQueued Status = iota
	StatusInProgress
	StatusFailed
	StatusCompleted
	StatusCanceled
)

type Status int

func (s *Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}
	*s = StatusFrom(statusStr)
	return nil
}

func (s *Status) Scan(src any) error {
	if val, ok := src.(string); ok {
		*s = StatusFrom(val)
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, Status(-2))
}

func (s *Status) Value() (driver.Value, error) {
	if s == nil {
		return nil, fmt.Errorf("unsupported Value, returing nil status as driver.Value")
	}
	if s.String() == "" {
		return nil, fmt.Errorf("unsupported Value, returing empty string status as driver.Value")
	}
	if s.String() == "unknown" {
		return nil, fmt.Errorf("unsupported Value, returing unknown status as driver.Value")
	}
	return int64(*s), nil
}

func (s *Status) String() string {
	switch *s {
	case StatusQueued:
		return "queued"
	case StatusInProgress:
		return "in_progress"
	case StatusFailed:
		return "failed"
	case StatusCompleted:
		return "completed"
	case StatusCanceled:
		return "canceled"
	case StatusUnset:
		return ""
	default:
		return "unknown"
	}
}

func StatusFrom(str string) Status {
	switch str {
	case "queued":
		return StatusQueued
	case "in_progress":
		return StatusInProgress
	case "failed":
		return StatusFailed
	case "completed":
		return StatusCompleted
	case "canceled":
		return StatusCanceled
	case "":
		return StatusUnset
	default:
		return -1
	}
}
