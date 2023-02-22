package migration

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

var tld = regexp.MustCompile(`\.(com|org)`)

type Request struct {
	ID          string    `json:"id,omitempty"`
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
	Request       `json:",inline"`
	Status        Status        `json:"status"`
	LastStep      Step          `json:"lastStep"`
	Builds        int           `json:"builds"`
	Releases      int           `json:"releases"`
	TotalDuration time.Duration `json:"totalDuration"`
	ErrorDetails  *string       `json:"errorDetails,omitempty"`
	QueuedAt      time.Time     `json:"queuedAt,omitempty"`
	UpdatedAt     time.Time     `json:"updatedAt,omitempty"`
}

func (t *Task) FromFQN() string {
	return fmt.Sprintf("%s/%s/%s", t.FromSource, t.FromOwner, t.FromName)
}

func (t *Task) ToFQN() string {
	return fmt.Sprintf("%s/%s/%s", t.ToSource, t.ToOwner, t.ToName)
}

func (t *Task) SqlArgs() []sql.NamedArg {
	return []sql.NamedArg{
		// id, status, lastStep
		sql.Named("id", t.ID),
		sql.Named("status", t.Status.String()),
		sql.Named("lastStep", t.LastStep.String()),
		sql.Named("builds", t.Builds),
		sql.Named("releases", t.Releases),
		sql.Named("totalDuration", t.TotalDuration),
		// From
		sql.Named("fromSource", t.FromSource),
		sql.Named("fromSourceName", tld.ReplaceAllString(t.FromSource, "")),
		sql.Named("fromOwner", t.FromOwner),
		sql.Named("fromName", t.FromName),
		sql.Named("fromFullName", t.FromOwner+"/"+t.FromName),
		// To
		sql.Named("toSource", t.ToSource),
		sql.Named("toSourceName", tld.ReplaceAllString(t.ToSource, "")),
		sql.Named("toOwner", t.ToOwner),
		sql.Named("toName", t.ToName),
		sql.Named("toFullName", t.ToOwner+"/"+t.ToName),
		// Other
		sql.Named("callbackURL", t.CallbackURL),
		sql.Named("errorDetails", t.ErrorDetails),
		sql.Named("queuedAt", t.QueuedAt),
		sql.Named("updatedAt", t.UpdatedAt),
	}
}
