package migration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Executor is a function that executes a migration task and returns any changes and if it succeeded.
type Executor func(ctx context.Context, task *Task) ([]Change, error)

// CallbackExecutor calls the callback URL if it's set.
func CallbackExecutor(_ context.Context, task *Task) ([]Change, error) {
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
}

// CompletedExecutor set Task.Status to StatusCompleted.
func CompletedExecutor(_ context.Context, task *Task) ([]Change, error) {
	task.Status = StatusCompleted
	return nil, nil
}