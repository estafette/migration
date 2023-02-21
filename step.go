package migration

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const (
	StepWaiting Step = iota
	StepReleasesFailed
	StepReleasesDone
	StepReleaseLogsFailed
	StepReleaseLogsDone
	StepReleaseLogObjectsFailed
	StepReleaseLogObjectsDone
	StepBuildsFailed
	StepBuildsDone
	StepBuildLogsFailed
	StepBuildLogsDone
	StepBuildLogObjectsFailed
	StepBuildLogObjectsDone
	StepBuildVersionsFailed
	StepBuildVersionsDone
	StepCallbackFailed
	StepCallbackDone
)

type Step int

func (s *Step) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Step) UnmarshalJSON(data []byte) error {
	var stepStr string
	if err := json.Unmarshal(data, &stepStr); err != nil {
		return err
	}
	*s = StepFrom(stepStr)
	return nil
}

func (s *Step) Scan(src any) error {
	if val, ok := src.(string); ok {
		*s = StepFrom(val)
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, Step(-1))
}

func (s *Step) Value() (driver.Value, error) {
	if s == nil {
		return nil, fmt.Errorf("unsupported Value, returing nil step as driver.Value")
	}
	if s.String() != "unknown" {
		return int64(*s), nil
	}
	return nil, fmt.Errorf("unsupported Value, returing unknown step as driver.Value")
}

func (s *Step) String() string {
	switch *s {
	case StepWaiting:
		return "waiting"
	case StepReleasesFailed:
		return "releases_failed"
	case StepReleasesDone:
		return "releases_done"
	case StepReleaseLogsFailed:
		return "release_logs_failed"
	case StepReleaseLogsDone:
		return "release_logs_done"
	case StepReleaseLogObjectsFailed:
		return "release_log_objects_failed"
	case StepReleaseLogObjectsDone:
		return "release_log_objects_done"
	case StepBuildsFailed:
		return "builds_failed"
	case StepBuildsDone:
		return "builds_done"
	case StepBuildLogsFailed:
		return "build_logs_failed"
	case StepBuildLogsDone:
		return "build_logs_done"
	case StepBuildLogObjectsFailed:
		return "build_log_objects_failed"
	case StepBuildLogObjectsDone:
		return "build_log_objects_done"
	case StepBuildVersionsFailed:
		return "build_versions_failed"
	case StepBuildVersionsDone:
		return "build_versions_done"
	case StepCallbackFailed:
		return "callback_failed"
	case StepCallbackDone:
		return "callback_done"
	default:
		return "unknown"
	}
}

func StepFrom(str string) Step {
	switch str {
	case "waiting":
		return StepWaiting
	case "releases_failed":
		return StepReleasesFailed
	case "releases_done":
		return StepReleasesDone
	case "release_logs_failed":
		return StepReleaseLogsFailed
	case "release_logs_done":
		return StepReleaseLogsDone
	case "release_log_objects_failed":
		return StepReleaseLogObjectsFailed
	case "release_log_objects_done":
		return StepReleaseLogObjectsDone
	case "builds_failed":
		return StepBuildsFailed
	case "builds_done":
		return StepBuildsDone
	case "build_logs_failed":
		return StepBuildLogsFailed
	case "build_logs_done":
		return StepBuildLogsDone
	case "build_log_objects_failed":
		return StepBuildLogObjectsFailed
	case "build_log_objects_done":
		return StepBuildLogObjectsDone
	case "build_versions_failed":
		return StepBuildVersionsFailed
	case "build_versions_done":
		return StepBuildVersionsDone
	case "callback_failed":
		return StepCallbackFailed
	case "callback_done":
		return StepCallbackDone
	default:
		return -1
	}
}
