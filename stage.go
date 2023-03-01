package migration

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type Stage interface {
	Name() StageName
	Success() Step
	Failure() Step
	Execute(ctx context.Context, task *Task) bool
}

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

// Execute stage and return changes and whether to stop execution
func (s *stage) Execute(ctx context.Context, task *Task) bool {
	start := time.Now()
	task.LastStep = s.Success()
	err := s.execute(ctx, task)
	if err != nil {
		task.Status = StatusFailed
		task.LastStep = s.Failure()
		errorDetails := err.Error()
		task.ErrorDetails = &errorDetails
		log.Error().Str("module", "github.com/estafette/migration").Err(err).Str("taskID", task.ID).Str("stage", string(s.Name())).Msg("stage failed")
		return false
	}
	// in update query duration is appended to existing value
	task.TotalDuration += time.Since(start)
	return true
}
