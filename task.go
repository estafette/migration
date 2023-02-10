package migration

import (
	"fmt"
	"time"
)

type TaskRequest struct {
	ID          string    `json:"id"`
	FromSource  string    `json:"fromSource"`
	FromOwner   string    `json:"fromOwner"`
	FromName    string    `json:"fromName"`
	ToSource    string    `json:"toSource"`
	ToOwner     string    `json:"toOwner"`
	ToName      string    `json:"toName"`
	CallbackURL *string   `json:"callbackURL,omitempty"`
	Restart     StageName `json:"restart,omitempty"`
}

type Task struct {
	TaskRequest   `json:",inline"`
	Status        Status        `json:"status"`
	LastStep      Step          `json:"lastStep"`
	Builds        int           `json:"builds"`
	Releases      int           `json:"releases"`
	TotalDuration time.Duration `json:"totalDuration"`
	ErrorDetails  *string       `json:"errorDetails,omitempty"`
	Duration      time.Duration
	QueuedAt      time.Time
	UpdatedAt     time.Time
	Changes       map[StageName][]Change
}

func (t *Task) FromFQN() string {
	return fmt.Sprintf("%s/%s/%s", t.FromSource, t.FromOwner, t.FromName)
}

func (t *Task) ToFQN() string {
	return fmt.Sprintf("%s/%s/%s", t.ToSource, t.ToOwner, t.ToName)
}
