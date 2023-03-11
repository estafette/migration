package migration

type Change struct {
	FromID int64
	ToID   int64
}

type Changes struct {
	Releases      int `json:"releases,omitempty"`
	ReleaseLogs   int `json:"releaseLogs,omitempty"`
	Builds        int `json:"builds,omitempty"`
	BuildLogs     int `json:"buildLogs,omitempty"`
	BuildVersions int `json:"buildVersions,omitempty"`
}
