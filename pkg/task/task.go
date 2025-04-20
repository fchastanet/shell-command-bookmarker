package task

import (
	"fmt"
	"time"

	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/tui"
)

type TaskSpecFunc func(id resource.ID) (TaskSpec, error)

type TaskSpec struct {
	// ID of the task to create.
	ID resource.ID
	// Short indicates whether the task is short-lived.
	Short bool
}

type Task struct {
	ID      resource.ID
	Created time.Time
	Updated time.Time
	Spec    TaskSpec
	Status  Status
	Short   bool
}

type Service struct {
}

type Factory struct {
	Counter   *int
	Program   string
	Publisher resource.Publisher[*Task]
}

// CreateTasks repeatedly invokes fn with each id in ids, creating a task for
// each invocation. If there is more than one id then a task group is created
// and the user sent to the task group's page; otherwise if only id is provided,
// the user is sent to the task's page.
func (h *Service) CreateTasks(
	fn TaskSpecFunc,
	ids ...resource.ID,
) func() (*Task, error) {
	return func() (*Task, error) {
		switch len(ids) {
		case 0:
			return nil, nil
		case 1:
			spec, err := fn(ids[0])
			if err != nil {
				return nil, fmt.Errorf("creating task: %w", err)
			}
			task, err := h.CreateTask(spec)
			if err != nil {
				return nil, fmt.Errorf("creating task: %w", err)
			}
			if task.Short {
				// Don't navigate the user to the task page for short tasks.
				return nil, nil
			}
			return NewNavigationMsg(tui.Kind(TaskKind), WithParent(task.ID))
		default:
			return nil, fmt.Errorf("task group not implemented yet")
		}
	}
}

// Create a task. The task is placed into a pending state and requires enqueuing
// before it'll be processed.
func (s *Service) CreateTask(spec TaskSpec) (*Task, error) {
	task, err := s.newTask(spec)
	if err != nil {
		return nil, err
	}

	s.logger.Info("created task", "task", task)

	// Add to db
	s.tasks.Add(task.ID, task)
	// Increment counter of number of live tasks
	*s.counter++

	if spec.AfterCreate != nil {
		spec.AfterCreate(task)
	}

	wait := make(chan error, 1)
	go func() {
		err := task.Wait()
		wait <- err
		if err != nil {
			s.logger.Error("task failed", "error", err, "task", task)
			return
		}
		s.logger.Info("completed task", "task", task)
	}()
	if spec.Wait {
		return task, <-wait
	}
	return task, nil
}
