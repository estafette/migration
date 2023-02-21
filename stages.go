package migration

import (
	"context"
	"github.com/rs/zerolog/log"
	"sort"
)

type Updater func(ctx context.Context, task *Task) error

type Stages interface {
	Current() Stage
	ExecuteNext(ctx context.Context, task *Task) ([]Change, bool)
	HasNext() bool
	Next() Stage
	Set(name StageName, executor Executor) Stages
}

type stages struct {
	current int
	stages  []*stage
	start   Step
	updater func(ctx context.Context, task *Task) error
}

// NewStages creates a new Stages instance which uses the given Updater to update the status of tasks.
func NewStages(updater Updater, start Step) Stages {
	return &stages{
		current: -1,
		updater: updater,
		start:   start,
	}
}

// Set the executor for the given stage name. If the stage does not exist it will be created.
// Order of this function's call does not matter, The stages are sorted by Step.
func (ss *stages) Set(name StageName, executor Executor) Stages {
	for index, s := range ss.stages {
		if s.Name() == name {
			ss.stages[index].execute = executor
			log.Debug().Msgf("github.com/estafette/migration: overriding stage %s", name)
			return ss
		}
	}
	ss.stages = append(ss.stages, &stage{
		name:    name,
		success: name.SuccessStep(),
		failure: name.FailedStep(),
		execute: executor,
	})
	log.Debug().Msgf("github.com/estafette/migration: appended stage %s", name)
	sort.Slice(ss.stages, func(i, j int) bool {
		return ss.stages[i].Failure() < ss.stages[j].Failure()
	})
	for i, s := range ss.stages {
		if ss.start < s.Failure() {
			break
		}
		ss.current = i - 1
		log.Info().Msgf("github.com/estafette/migration: skipping stage %s\n", s.Name())
	}
	return ss
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
	return ss.stages[ss.current]
}

// Current returns the current stage or nil if there is no current stage.
func (ss *stages) Current() Stage {
	if ss.current == -1 || ss.current >= len(ss.stages) {
		return nil
	}
	return ss.stages[ss.current]
}

// ExecuteNext executes the next stage, saves result to using Updater and returns the changes and if the stage failed.
func (ss *stages) ExecuteNext(ctx context.Context, task *Task) (change []Change, failed bool) {
	stg := ss.Next()
	log.Info().Str("taskID", task.ID).Str("stage", string(stg.Name())).Msg("github.com/estafette/migration: stage started")
	start := task.TotalDuration
	change, failed = stg.Execute(ctx, task)
	err := ss.updater(ctx, task)
	if err != nil {
		log.Error().Err(err).Str("taskID", task.ID).Msg("github.com/estafette/migration: error updating migration status")
		return
	}
	if failed {
		log.Info().
			Str("taskID", task.ID).Str("fromFQN", task.FromFQN()).Str("toFQN", task.ToFQN()).Str("stage", string(stg.Name())).
			Msg("task failed, stopping migration")
		return
	}
	log.Info().
		Dur("took", task.TotalDuration-start).
		Str("taskID", task.ID).Str("fromFQN", task.FromFQN()).Str("toFQN", task.ToFQN()).Str("stage", string(stg.Name())).
		Msg("github.com/estafette/migration: stage done")
	return
}
