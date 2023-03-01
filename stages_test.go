package migration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func _WaitingTask() *Task {
	return &Task{
		Request:  Request{ID: "test-1", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
		Status:   StatusQueued,
		LastStep: StepWaiting,
	}
}
func _RestartedTask() *Task {
	return &Task{
		Request:  Request{ID: "test-1", FromSource: "github.com", FromOwner: "estafette", FromName: "migration", ToSource: "github.com", ToOwner: "estafette_new", ToName: "migration_new"},
		Status:   StatusQueued,
		LastStep: StepBuildsFailed,
	}
}

type mockUpdater struct {
	mock.Mock
}

func (m *mockUpdater) update(ctx context.Context, task *Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func TestStages(t *testing.T) {
	mockedUpdater := &mockUpdater{}
	mockedUpdater.On("update", mock.Anything, mock.Anything).Return(nil)
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(nil)
	ss := NewStages(mockedUpdater.update, _WaitingTask())
	cont := t.Run("Set_and_Len", func(t *testing.T) {
		ss.
			Set(ReleasesStage, mockedExecutor.execute).
			Set(ReleaseLogsStage, mockedExecutor.execute).
			Set(ReleaseLogObjectsStage, mockedExecutor.execute).
			Set(BuildsStage, mockedExecutor.execute).
			Set(BuildLogsStage, mockedExecutor.execute).
			Set(BuildLogObjectsStage, mockedExecutor.execute).
			Set(BuildVersionsStage, mockedExecutor.execute).
			Set(CallbackStage, mockedExecutor.execute)
		assert.Equal(t, 8, ss.Len())
	})
	if !cont {
		return
	}
	cont = t.Run("HasNext", func(t *testing.T) { assert.True(t, ss.HasNext()) })
	if !cont {
		return
	}
	cont = t.Run("ExecuteNext", func(t *testing.T) {
		for ss.HasNext() {
			failed := ss.ExecuteNext(context.TODO())
			assert.False(t, failed)
		}
	})
	if !cont {
		return
	}
	t.Run("Current", func(t *testing.T) { assert.Equal(t, CallbackStage, ss.Current().Name()) })
	mockedExecutor.AssertNumberOfCalls(t, "execute", 8)
	mockedUpdater.AssertNumberOfCalls(t, "update", 8)
}

func TestStages_Skipped(t *testing.T) {
	mockedUpdater := &mockUpdater{}
	mockedUpdater.On("update", mock.Anything, mock.Anything).Return(nil)
	skippedExecutor := &mockExecutor{}
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(nil)
	ss := NewStages(mockedUpdater.update, _RestartedTask()).
		Set(ReleasesStage, skippedExecutor.execute).
		Set(ReleaseLogsStage, skippedExecutor.execute).
		Set(ReleaseLogObjectsStage, skippedExecutor.execute).
		Set(BuildsStage, mockedExecutor.execute).
		Set(BuildLogsStage, mockedExecutor.execute).
		Set(BuildLogObjectsStage, mockedExecutor.execute).
		Set(BuildVersionsStage, mockedExecutor.execute).
		Set(CallbackStage, mockedExecutor.execute)
	for ss.HasNext() {
		failed := ss.ExecuteNext(context.TODO())
		assert.False(t, failed)
	}
	t.Run("Current", func(t *testing.T) { assert.Equal(t, CallbackStage, ss.Current().Name()) })
	mockedExecutor.AssertNumberOfCalls(t, "execute", 5)
	mockedUpdater.AssertNumberOfCalls(t, "update", 5)
	skippedExecutor.AssertNotCalled(t, "execute", mock.Anything, mock.Anything)
}
