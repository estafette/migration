package migration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallbackExecutor(t *testing.T) {
	called := false
	shouldBe := assert.New(t)
	var testServer *httptest.Server
	testServer = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		called = true
		res.WriteHeader(http.StatusOK)
		shouldBe.Equal("application/json", req.Header.Get("Content-Type"))
		data, err := io.ReadAll(req.Body)
		shouldBe.Nil(err)
		expected := fmt.Sprintf(`{"id":"test-456","fromSource":"x","fromOwner":"y","fromName":"z","toSource":"a","toOwner":"b","toName":"c","callbackURL":"%s","status":"in_progress","lastStep":"build_versions_done","builds":10,"releases":11,"totalDuration":12,"queuedAt":"2020-01-01T00:00:00Z","updatedAt":"2020-01-01T00:00:00Z"}`, testServer.URL)
		shouldBe.Equal(expected, string(data))
	}))
	defer func() { testServer.Close() }()
	err := CallbackExecutor(context.TODO(), &Task{
		Request:       Request{ID: "test-456", FromSource: "x", FromOwner: "y", FromName: "z", ToSource: "a", ToOwner: "b", ToName: "c", CallbackURL: &testServer.URL},
		Status:        StatusInProgress,
		LastStep:      StepBuildVersionsDone,
		Builds:        10,
		Releases:      11,
		TotalDuration: 12,
		QueuedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	shouldBe.Nil(err)
	shouldBe.True(called)
}

func TestCompletedExecutor(t *testing.T) {
	shouldBe := assert.New(t)
	task := &Task{Status: StatusInProgress}
	err := CompletedExecutor(context.TODO(), task)
	shouldBe.Nil(err)
	shouldBe.Equal(StatusCompleted, task.Status)
}
