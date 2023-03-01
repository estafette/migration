package migration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockExecutor struct {
	mock.Mock
}

func (m *mockExecutor) execute(ctx context.Context, task *Task) error {
	time.Sleep(50 * time.Millisecond)
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

func TestStage_Execute_Success(t *testing.T) {
	shouldBe := assert.New(t)
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(nil).Once()
	s := &stage{
		name:    "test",
		success: Step(0),
		failure: Step(1),
		execute: mockedExecutor.execute,
	}
	task := &Task{Request: Request{ID: "test-123"}}
	failed := s.Execute(context.TODO(), task)
	mockedExecutor.AssertExpectations(t)
	shouldBe.False(failed)
	shouldBe.Equal(Step(0), task.LastStep)
	shouldBe.GreaterOrEqual(task.TotalDuration, 50*time.Millisecond)
}

func TestStage_Execute_Failure(t *testing.T) {
	shouldBe := assert.New(t)
	expected := "test error"
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(errors.New(expected)).Once()
	s := &stage{
		name:    "test",
		success: Step(0),
		failure: Step(1),
		execute: mockedExecutor.execute,
	}
	task := &Task{Request: Request{ID: "test-123"}}
	failed := s.Execute(context.TODO(), task)
	mockedExecutor.AssertExpectations(t)
	shouldBe.True(failed)
	shouldBe.Equal(Step(1), task.LastStep)
	shouldBe.Equal(time.Duration(0), task.TotalDuration)
	shouldBe.Equal(&expected, task.ErrorDetails)
	shouldBe.Equal(StatusFailed, task.Status)
}
