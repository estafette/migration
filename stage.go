package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const (
	LastStage              StageName = "last_stage"
	ReleasesStage          StageName = "releases"
	ReleaseLogsStage       StageName = "release_logs"
	ReleaseLogObjectsStage StageName = "release_log_objects"
	BuildsStage            StageName = "builds"
	BuildLogsStage         StageName = "build_logs"
	BuildLogObjectsStage   StageName = "build_log_objects"
	BuildVersionsStage     StageName = "build_versions"
	CallbackStage          StageName = "callback"
	CompletedStage         StageName = "completed"
)

type StageName string

type Stage interface {
	Name() StageName
	Success() Step
	Failure() Step
	Execute(ctx context.Context, task *Task) ([]Change, bool)
}

type Executor func(ctx context.Context, task *Task) ([]Change, error)

type stage struct {
	name    StageName
	success Step
	failure Step
	execute Executor
}

func (s *stage) Name() StageName {
	return s.name
}

func (s *stage) Success() Step {
	return s.success
}

func (s *stage) Failure() Step {
	return s.failure
}

func (s *stage) Execute(ctx context.Context, task *Task) ([]Change, bool) {
	start := time.Now()
	changes, err := s.execute(ctx, task)
	task.LastStep = s.Success()
	if err != nil {
		task.LastStep = s.Failure()
		errorDetails := err.Error()
		task.ErrorDetails = &errorDetails
		log.Error().Err(err).Msgf("migration stage %s: execution failed", s.name)
		return nil, true
	}
	// in update query duration is appended to existing value
	task.TotalDuration += time.Since(start)
	return changes, false
}

func Releases(fn Executor) Stage {
	return &stage{
		name:    ReleasesStage,
		success: StepReleasesDone,
		failure: StepReleasesFailed,
		execute: func(ctx context.Context, task *Task) ([]Change, error) {
			changes, err := fn(ctx, task)
			if err == nil {
				task.Releases = len(changes)
			}
			return changes, err
		},
	}
}

func ReleaseLogs(fn Executor) Stage {
	return &stage{
		name:    ReleaseLogsStage,
		success: StepReleaseLogsDone,
		failure: StepReleaseLogsFailed,
		execute: fn,
	}
}

func ReleaseLogObjects(fn Executor) Stage {
	return &stage{
		name:    ReleaseLogObjectsStage,
		success: StepReleaseLogObjectsDone,
		failure: StepReleaseLogObjectsFailed,
		execute: fn,
	}
}

func Builds(fn Executor) Stage {
	return &stage{
		name:    BuildsStage,
		success: StepBuildsDone,
		failure: StepBuildsFailed,
		execute: func(ctx context.Context, task *Task) ([]Change, error) {
			changes, err := fn(ctx, task)
			if err == nil {
				task.Builds = len(changes)
			}
			return changes, err
		},
	}
}

func BuildLogs(fn Executor) Stage {
	return &stage{
		name:    BuildLogsStage,
		success: StepBuildLogsDone,
		failure: StepBuildLogsFailed,
		execute: fn,
	}
}

func BuildLogObjects(fn Executor) Stage {
	return &stage{
		name:    BuildLogObjectsStage,
		success: StepBuildLogObjectsDone,
		failure: StepBuildLogObjectsFailed,
		execute: fn,
	}
}

func BuildVersions(fn Executor) Stage {
	return &stage{
		name:    BuildVersionsStage,
		success: StepBuildVersionsDone,
		failure: StepBuildVersionsFailed,
		execute: fn,
	}
}

func Callback() Stage {
	return &stage{
		name:    CallbackStage,
		success: StepCallbackDone,
		failure: StepCallbackFailed,
		execute: func(ctx context.Context, task *Task) ([]Change, error) {
			if task.CallbackURL == nil {
				payload, err := json.Marshal(task)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal migration callback payload: %w", err)
				}
				resp, err := http.Post(*task.CallbackURL, "application/json", bytes.NewBuffer(payload))
				if err != nil {
					return nil, fmt.Errorf("failed to httpPost migration callback: %w", err)
				}
				if resp.StatusCode <= http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
					return nil, fmt.Errorf("migration callback returned invalid status code %d", resp.StatusCode)
				}
			}
			return nil, nil
		},
	}
}

func Completed() Stage {
	return &stage{
		name:    CompletedStage,
		success: -1,
		failure: -1,
		execute: func(ctx context.Context, task *Task) ([]Change, error) {
			task.Status = StatusCompleted
			return nil, nil
		},
	}
}

func FailedStepOf(stageName StageName) Step {
	switch stageName {
	case ReleasesStage:
		return StepReleasesFailed
	case ReleaseLogsStage:
		return StepReleaseLogsFailed
	case ReleaseLogObjectsStage:
		return StepReleaseLogObjectsFailed
	case BuildsStage:
		return StepBuildsFailed
	case BuildLogsStage:
		return StepBuildLogsFailed
	case BuildLogObjectsStage:
		return StepBuildLogObjectsFailed
	case BuildVersionsStage:
		return StepBuildVersionsFailed
	case CallbackStage:
		return StepCallbackFailed
	case CompletedStage: // special case considering Callback is last step
		return StepCallbackDone
	default:
		return StepWaiting
	}
}
