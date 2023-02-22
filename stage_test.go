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

func (m *mockExecutor) execute(ctx context.Context, task *Task) ([]Change, error) {
	time.Sleep(50 * time.Millisecond)
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Change), args.Error(1)
}

func TestStage_Execute_Success(t *testing.T) {
	shouldBe := assert.New(t)
	expected := []Change{{FromID: 123, ToID: 456}}
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(expected, nil).Once()
	s := &stage{
		name:    "test",
		success: Step(0),
		failure: Step(1),
		execute: mockedExecutor.execute,
	}
	task := &Task{Request: Request{ID: "test-123"}}
	actual, failed := s.Execute(context.TODO(), task)
	mockedExecutor.AssertExpectations(t)
	shouldBe.False(failed)
	shouldBe.Equal(expected, actual)
	shouldBe.Equal(Step(0), task.LastStep)
	shouldBe.GreaterOrEqual(task.TotalDuration, 50*time.Millisecond)
}

func TestStage_Execute_Failure(t *testing.T) {
	shouldBe := assert.New(t)
	expected := "test error"
	mockedExecutor := &mockExecutor{}
	mockedExecutor.On("execute", mock.Anything, mock.Anything).Return(nil, errors.New(expected)).Once()
	s := &stage{
		name:    "test",
		success: Step(0),
		failure: Step(1),
		execute: mockedExecutor.execute,
	}
	task := &Task{Request: Request{ID: "test-123"}}
	actual, failed := s.Execute(context.TODO(), task)
	mockedExecutor.AssertExpectations(t)
	shouldBe.True(failed)
	shouldBe.Nil(actual)
	shouldBe.Equal(Step(1), task.LastStep)
	shouldBe.Equal(time.Duration(0), task.TotalDuration)
	shouldBe.Equal(&expected, task.ErrorDetails)
}
