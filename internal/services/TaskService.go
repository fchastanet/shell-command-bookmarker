package services

import (
	"slices"

	"github.com/fchastanet/shell-command-bookmarker/pkg/logging"
	"github.com/fchastanet/shell-command-bookmarker/pkg/pubsub"
	"github.com/fchastanet/shell-command-bookmarker/pkg/resource"
	"github.com/fchastanet/shell-command-bookmarker/pkg/task"
)

type taskServiceInterface interface {
	Enqueue(taskID resource.ID) (*task.Task, error)
	List(opts ListOptions) []*task.Task
	Get(taskID resource.ID) (*task.Task, error)
	Cancel(taskID resource.ID) (*task.Task, error)
	Delete(taskID resource.ID) error
	Counter() int
}

type taskFactoryInterface interface {
	CreateTasks(
		fn task.TaskSpecFunc,
		ids ...resource.ID,
	) func() (*task.Task, error)
}

type Service struct {
	tasks       *resource.Table[*task.Task]
	counter     *int
	logger      logging.Interface
	taskService taskServiceInterface
	factory     taskFactoryInterface

	TaskBroker *pubsub.Broker[*task.Task]
}

type ServiceOptions struct {
	Program string
	Logger  logging.Interface
}

func NewService(opts ServiceOptions) *Service {
	var counter int

	taskBroker := pubsub.NewBroker[*task.Task](opts.Logger)

	factory := &task.Factory{
		Counter:   &counter,
		Program:   opts.Program,
		Publisher: taskBroker,
	}

	return &Service{
		tasks:      resource.NewTable(taskBroker),
		TaskBroker: taskBroker,
		factory:    factory,
		counter:    &counter,
		logger:     opts.Logger,
	}
}

// Enqueue moves the task onto the global queue for processing.
func (s *Service) Enqueue(taskID resource.ID) (*task.Task, error) {
	task, err := s.tasks.Update(taskID, func(existing *task.Task) error {
		existing.updateState(Queued)
		return nil
	})
	if err != nil {
		s.logger.Error("enqueuing task", "error", err)
		return nil, err
	}
	s.logger.Debug("enqueued task", "task", task)
	return task, nil
}

type ListOptions struct {
	// Filter tasks by those with a matching module path. Optional.
	Path *string
	// Filter tasks by status: match task if it has one of these statuses.
	// Optional.
	Status []Status
	// Order tasks by oldest first (true), or newest first (false)
	Oldest bool
	// Filter tasks by only those that are blocking. If false, both blocking and
	// non-blocking tasks are returned.
	Blocking bool
	// Only return those tasks that are exclusive. If false, both exclusive and
	// non-exclusive tasks are returned.
	Exclusive bool
}

type taskLister interface {
	List(opts ListOptions) []*task.Task
}

func (s *Service) CreateTasks(
	fn task.TaskSpecFunc,
	ids ...resource.ID,
) func() (*task.Task, error) {
	return s.factory.CreateTasks(fn, ids...)
}

func (s *Service) List(opts ListOptions) []*task.Task {
	tasks := s.tasks.List()

	// Filter list according to options
	var i int
	for _, t := range tasks {
		if opts.Path != nil && *opts.Path != t.Path {
			continue
		}
		if opts.Status != nil {
			if !slices.Contains(opts.Status, t.State) {
				continue
			}
		}
		tasks[i] = t
		i++
	}
	tasks = tasks[:i]

	// Sort list according to options
	slices.SortFunc(tasks, func(a, b *task.Task) int {
		cmp := a.Updated.Compare(b.Updated)
		if opts.Oldest {
			return cmp
		}
		return -cmp
	})

	return tasks
}

func (s *Service) Get(taskID resource.ID) (*task.Task, error) {
	return s.tasks.Get(taskID)
}

func (s *Service) Cancel(taskID resource.ID) (*task.Task, error) {
	task, err := func() (*task.Task, error) {
		task, err := s.tasks.Get(taskID)
		if err != nil {
			return nil, err
		}
		return task, task.Cancel()
	}()
	if err != nil {
		s.logger.Error("canceling task", "id", taskID, "error", err)
		return nil, err
	}

	s.logger.Info("canceled task", "task", task)
	return task, nil
}

func (s *Service) Delete(taskID resource.ID) error {
	// TODO: only allow deleting task if in finished state (error message should
	// instruct user to cancel task first).
	s.tasks.Delete(taskID)
	return nil
}

func (s *Service) Counter() int {
	return *s.counter
}
