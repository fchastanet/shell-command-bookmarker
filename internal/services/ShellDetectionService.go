package services

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// ShellType represents the type of shell
type ShellType string

const (
	// ShellTypeUnknown represents an unknown shell
	ShellTypeUnknown ShellType = "unknown"
	// ShellTypeBash represents the Bash shell
	ShellTypeBash ShellType = "bash"
	// ShellTypeZsh represents the Zsh shell
	ShellTypeZsh ShellType = "zsh"

	OSLinux   = "linux"
	OSDarwin  = "darwin"
	OSFreeBSD = "freebsd"
	OSOpenBSD = "openbsd"
	OSWindows = "windows"
	OSUnknown = "unknown"
)

// ShellDetectionService provides methods to detect the current shell
type ShellDetectionService struct {
	execCommand               ExecCommandFunc
	getPpid                   func() int
	getOS                     func() string
	readFile                  func(name string) ([]byte, error)
	getLinuxProcessName       func(pid int, readFile ReadFileFunc) (string, error)
	getLinuxLegacyProcessName func(pid int, readFile ReadFileFunc) (string, error)
	getOtherOsProcessName     func(pid int, execCommand ExecCommandFunc) (string, error)
}

type ShellDetectionServiceInterface interface {
	DetectShell() ShellType
}

type (
	ReadFileFunc    func(name string) ([]byte, error)
	ExecCommandFunc func(name string, arg ...string) *exec.Cmd
)

// NewShellDetectionService creates a new ShellDetectionService
func NewShellDetectionService() *ShellDetectionService {
	return &ShellDetectionService{
		execCommand:               exec.Command,
		getPpid:                   os.Getppid,
		getOS:                     func() string { return runtime.GOOS },
		readFile:                  os.ReadFile,
		getLinuxProcessName:       getLinuxProcessName,
		getLinuxLegacyProcessName: getLinuxLegacyProcessName,
		getOtherOsProcessName:     getOtherOsProcessName,
	}
}

func getLinuxProcessName(pid int, readFile ReadFileFunc) (string, error) {
	content, err := readFile(fmt.Sprintf("/proc/%d/comm", pid))
	if err == nil {
		return strings.TrimSpace(string(content)), nil
	}
	return "", err
}

func getLinuxLegacyProcessName(pid int, readFile ReadFileFunc) (string, error) {
	content, err := readFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err == nil {
		// cmdline is null-byte separated
		parts := strings.Split(string(content), "\x00")
		if len(parts) > 0 && parts[0] != "" {
			return filepath.Base(parts[0]), nil
		}
	}
	return "", err
}

func getOtherOsProcessName(pid int, execCommand ExecCommandFunc) (string, error) {
	cmd := execCommand("ps", "-o", "comm=", "-p", strconv.Itoa(pid))
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}
	return "", err
}

// DetectShell attempts to detect if the parent process is bash or zsh
func (s *ShellDetectionService) DetectShell() ShellType {
	// Check parent process name
	shellType := s.detectParentProcessShell()
	if shellType != ShellTypeUnknown {
		return shellType
	}

	// Fallback to default
	slog.Warn("Could not detect shell type")
	return ShellTypeUnknown
}

func (s *ShellDetectionService) detectParentProcessShell() ShellType {
	// Early return for unsupported OS
	if s.getOS() == OSWindows {
		return ShellTypeUnknown
	}

	ppid := s.getPpid()

	// Try OS-specific detection
	var processName string

	switch s.getOS() {
	case OSLinux:
		processName = s.detectLinuxShell(ppid)
	case OSDarwin, OSFreeBSD, OSOpenBSD:
		processName = s.detectUnixShell(ppid)
	}

	return s.convertProcessNameToShellType(processName, ppid)
}

// Helper method for Linux shell detection
func (s *ShellDetectionService) detectLinuxShell(pid int) string {
	// Try /proc/pid/comm first
	if name, err := s.getLinuxProcessName(pid, s.readFile); err == nil && name != "" {
		return name
	}

	// Fall back to /proc/pid/cmdline
	if name, err := s.getLinuxLegacyProcessName(pid, s.readFile); err == nil && name != "" {
		return name
	}

	return ""
}

// Helper method for macOS/BSD shell detection
func (s *ShellDetectionService) detectUnixShell(pid int) string {
	name, err := s.getOtherOsProcessName(pid, s.execCommand)
	if err == nil && name != "" {
		return name
	}
	return ""
}

func (*ShellDetectionService) convertProcessNameToShellType(processName string, ppid int) ShellType {
	slog.Debug("Detected parent process", "name", processName, "pid", ppid)

	switch {
	case strings.Contains(processName, string(ShellTypeBash)):
		return ShellTypeBash
	case strings.Contains(processName, string(ShellTypeZsh)):
		return ShellTypeZsh
	default:
		return ShellTypeUnknown
	}
}
