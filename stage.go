package migration

import (
	"context"
	"github.com/rs/zerolog/log"
	"time"
)

type Stage interface {
	Name() StageName
	Success() Step
	Failure() Step
	Execute(ctx context.Context, task *Task) ([]Change, bool)
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
func (s *stage) Execute(ctx context.Context, task *Task) ([]Change, bool) {
	start := time.Now()
	changes, err := s.execute(ctx, task)
	task.LastStep = s.Success()
	if err != nil {
		task.LastStep = s.Failure()
		errorDetails := err.Error()
		task.ErrorDetails = &errorDetails
		log.Error().Err(err).Str("taskID", task.ID).Str("stage", string(s.Name())).Msg("github.com/estafette/migration: stage failed")
		return nil, true
	}
	// in update query duration is appended to existing value
	task.TotalDuration += time.Since(start)
	return changes, false
}
