package migration

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTask_SqlArgs(t *testing.T) {
	callback := "http://localhost:8080"
	errorDetails := "test error message"
	task := &Task{
		Request: Request{
			ID:          "1234",
			FromSource:  "github.com",
			FromOwner:   "foo",
			FromName:    "bar",
			ToSource:    "gitlab.com",
			ToOwner:     "foo",
			ToName:      "baz",
			CallbackURL: &callback,
			Restart:     "builds",
		},
		Status:        StatusQueued,
		LastStep:      StepBuildsFailed,
		Builds:        0,
		Releases:      100,
		TotalDuration: 101,
		ErrorDetails:  &errorDetails,
		QueuedAt:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	args := task.SqlArgs()
	shouldbe := assert.New(t)
	shouldbe.Equal(20, len(args))
	shouldbe.Equal([]sql.NamedArg{
		sql.Named("updatedAt", task.UpdatedAt),
		sql.Named("totalDuration", task.TotalDuration),
		sql.Named("toSourceName", tld.ReplaceAllString(task.ToSource, "")),
		sql.Named("toSource", task.ToSource),
		sql.Named("toOwner", task.ToOwner),
		sql.Named("toName", task.ToName),
		sql.Named("toFullName", task.ToOwner+"/"+task.ToName),
		sql.Named("status", task.Status.String()),
		sql.Named("releases", task.Releases),
		sql.Named("queuedAt", task.QueuedAt),
		sql.Named("lastStep", task.LastStep.String()),
		sql.Named("id", task.ID),
		sql.Named("fromSourceName", tld.ReplaceAllString(task.FromSource, "")),
		sql.Named("fromSource", task.FromSource),
		sql.Named("fromOwner", task.FromOwner),
		sql.Named("fromName", task.FromName),
		sql.Named("fromFullName", task.FromOwner+"/"+task.FromName),
		sql.Named("errorDetails", task.ErrorDetails),
		sql.Named("callbackURL", task.CallbackURL),
		sql.Named("builds", task.Builds),
	}, args)
}
