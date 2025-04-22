package command

// commandReloadedMsg is sent when a command reload has finished.
type commandReloadedMsg struct {
	err error
}
