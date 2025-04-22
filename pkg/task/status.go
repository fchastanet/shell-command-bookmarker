package task

import "time"

// Status is a stage in the lifecycle of a task.
type Status string

const (
	Pending  Status = "pending"
	Queued   Status = "queued"
	Running  Status = "running"
	Exited   Status = "exited"
	Errored  Status = "errored"
	Canceled Status = "canceled"

	MaxStatusLen = len(Canceled)
)

// IsFinal returns true if the state is a final state.
func (s Status) IsFinal() bool {
	switch s {
	case Errored, Exited, Canceled:
		return true
	case Pending, Queued, Running:
		return false
	default:
		return false
	}
}

func (s Status) String() string {
	switch s {
	case Pending:
		return "pending"
	case Queued:
		return "queued"
	case Running:
		return "running"
	case Exited:
		return "exited"
	case Errored:
		return "errored"
	case Canceled:
		return "canceled"
	default:
		return "unknown"
	}
}

type statusTimestamps struct {
	started time.Time
	ended   time.Time
}

func (sd statusTimestamps) Elapsed() time.Duration {
	if sd.started.IsZero() {
		return 0
	}
	if sd.ended.IsZero() {
		return time.Since(sd.started)
	}
	return sd.ended.Sub(sd.started)
}
