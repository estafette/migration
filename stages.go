package migration

import (
	"context"
	"sort"

	"github.com/rs/zerolog/log"
)

type Updater func(ctx context.Context, task *Task) error

type Stages interface {
	Current() Stage
	ExecuteNext(ctx context.Context) ([]Change, bool)
	HasNext() bool
	Len() int
	Set(name StageName, executor Executor) Stages
}

type stages struct {
	current int
	task    *Task
	stages  []*stage
	updater func(ctx context.Context, task *Task) error
}

// NewStages creates a new Stages instance which uses the given Updater to update the status of tasks.
func NewStages(updater Updater, task *Task) Stages {
	return &stages{
		current: -1,
		updater: updater,
		task:    task,
	}
}

// Current returns the current stage or nil if there is no current stage.
func (ss *stages) Current() Stage {
	if ss.current == -1 || ss.current >= len(ss.stages) {
		return nil
	}
	return ss.stages[ss.current]
}

// ExecuteNext executes the next stage, saves result to using Updater and returns the changes and if the stage failed.
func (ss *stages) ExecuteNext(ctx context.Context) ([]Change, bool) {
	defer ss.updateStatus(ctx)
	stg := ss.Next()
	log.Info().Str("module", "github.com/estafette/migration").Str("taskID", ss.task.ID).Str("stage", string(stg.Name())).Msg("stage started")
	start := ss.task.TotalDuration
	changes, failed := stg.Execute(ctx, ss.task)
	if failed {
		log.Warn().Str("module", "github.com/estafette/migration").Str("taskID", ss.task.ID).Msg("task failed, stopping migration")
		return changes, failed
	}
	log.Info().Str("module", "github.com/estafette/migration").Dur("took", ss.task.TotalDuration-start).Str("taskID", ss.task.ID).Msg("stage done")
	return changes, failed
}

// HasNext returns true if there is a next stage.
func (ss *stages) HasNext() bool {
	return ss.current+1 < len(ss.stages)
}

// Next returns the next stage or nil if there is no next stage.
func (ss *stages) Next() Stage {
	if !ss.HasNext() {
		return nil
	}
	ss.current++
	return ss.Current()
}

func (ss *stages) Len() int {
	return len(ss.stages)
}

// Set the executor for the given stage name. If the stage is before Task.LastStep it will not be added.
// Multiple calls to this function can be unordered, the stages are executed in ascending order of Step.
func (ss *stages) Set(name StageName, executor Executor) Stages {
	if ss.task.LastStep > name.SuccessStep() {
		log.Info().Str("module", "github.com/estafette/migration").Msgf("not adding stage %s", name)
		return ss
	}
	defer sort.Slice(ss.stages, func(i, j int) bool {
		return ss.stages[i].Failure() < ss.stages[j].Failure()
	})
	for index, s := range ss.stages {
		if s.Name() == name {
			log.Warn().Str("module", "github.com/estafette/migration").Msgf("overriding existing stage %s", name)
			ss.stages[index].execute = executor
			return ss
		}
	}
	log.Debug().Str("module", "github.com/estafette/migration").Msgf("appended stage %s", name)
	ss.stages = append(ss.stages, &stage{
		name:    name,
		success: name.SuccessStep(),
		failure: name.FailedStep(),
		execute: executor,
	})
	return ss
}

func (ss *stages) updateStatus(ctx context.Context) {
	if err := ss.updater(ctx, ss.task); err != nil {
		log.Error().Str("module", "github.com/estafette/migration").Err(err).Str("taskID", ss.task.ID).Msg("error updating migration status")
	}
}
