package migration

const (
	LastStage              StageName = "last_stage"
	ReleasesStage          StageName = "releases"
	ReleaseLogsStage       StageName = "release_logs"
	ReleaseLogObjectsStage StageName = "release_log_objects"
	BuildsStage            StageName = "builds"
	BuildLogsStage         StageName = "build_logs"
	BuildLogObjectsStage   StageName = "build_log_objects"
	BuildVersionsStage     StageName = "build_versions"
	CallbackStage          StageName = "callback"
	CompletedStage         StageName = "completed"
)

type StageName string

// SuccessStep for this stage name
func (sn StageName) SuccessStep() Step {
	switch sn {
	case ReleasesStage:
		return StepReleasesDone
	case ReleaseLogsStage:
		return StepReleaseLogsDone
	case ReleaseLogObjectsStage:
		return StepReleaseLogObjectsDone
	case BuildsStage:
		return StepBuildsDone
	case BuildLogsStage:
		return StepBuildLogsDone
	case BuildLogObjectsStage:
		return StepBuildLogObjectsDone
	case BuildVersionsStage:
		return StepBuildVersionsDone
	case CallbackStage:
		return StepCallbackDone
	case CompletedStage: // special case considering Callback is last step
		return StepCompletionDone
	default:
		panic("unknown stage name " + string(sn))
	}
}

// FailedStep for this stage name
func (sn StageName) FailedStep() Step {
	switch sn {
	case ReleasesStage:
		return StepReleasesFailed
	case ReleaseLogsStage:
		return StepReleaseLogsFailed
	case ReleaseLogObjectsStage:
		return StepReleaseLogObjectsFailed
	case BuildsStage:
		return StepBuildsFailed
	case BuildLogsStage:
		return StepBuildLogsFailed
	case BuildLogObjectsStage:
		return StepBuildLogObjectsFailed
	case BuildVersionsStage:
		return StepBuildVersionsFailed
	case CallbackStage:
		return StepCallbackFailed
	case CompletedStage:
		return StepCompletionFailed
	default:
		panic("unknown stage name " + string(sn))
	}
}
