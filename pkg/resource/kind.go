package resource

type Kind int

const (
	Global Kind = iota
	Command
	Task
)

func (k Kind) String() string {
	return [...]string{
		"global",
		"command",
		"task",
	}[k]
}
