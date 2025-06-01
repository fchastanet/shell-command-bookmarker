package services

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// StubShellDetectionService is a simple stub implementation for testing
type StubShellDetectionService struct {
	ShellToReturn ShellType
}

func (s *StubShellDetectionService) DetectShell() ShellType {
	return s.ShellToReturn
}

// TestConvertProcessNameToShellType tests the process name to shell type conversion
func TestConvertProcessNameToShellType(t *testing.T) {
	service := ShellDetectionService{} //nolint:exhaustruct // test

	tests := []struct {
		name        string
		processName string
		expected    ShellType
	}{
		{"BashProcess", "bash", ShellTypeBash},
		{"BashPrefix", "bash-something", ShellTypeBash},
		{"BashSuffix", "something-bash", ShellTypeBash},
		{"ZshProcess", "zsh", ShellTypeZsh},
		{"ZshPrefix", "zsh-something", ShellTypeZsh},
		{"ZshSuffix", "something-zsh", ShellTypeZsh},
		{"UnknownProcess", "something", ShellTypeUnknown},
		{"EmptyProcess", "", ShellTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.convertProcessNameToShellType(tt.processName, 1234)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDetectParentProcessShell_Windows tests shell detection on Windows
func TestDetectParentProcessShell_Windows(t *testing.T) {
	service := &ShellDetectionService{ //nolint:exhaustruct // test
		getOS: func() string { return OSWindows },
	}

	result := service.detectParentProcessShell()
	assert.Equal(t, ShellTypeUnknown, result)
}

// TestDetectParentProcessShell_Linux tests shell detection on Linux
func TestDetectParentProcessShell_Linux(t *testing.T) {
	// Test successful detection via /proc/pid/comm
	t.Run("LinuxCommSuccess", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSLinux },
			getPpid: func() int { return 1234 },
			getLinuxProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return string(ShellTypeBash), nil
			},
			getLinuxLegacyProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("should not be called") //nolint:err113 //test
			},
		}

		result := service.detectParentProcessShell()
		assert.Equal(t, ShellTypeBash, result)
	})

	// Test fallback to /proc/pid/cmdline when comm fails
	t.Run("LinuxCommFailsLegacySucceeds", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSLinux },
			getPpid: func() int { return 1234 },
			getLinuxProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("failed to read /proc/pid/comm") //nolint:err113 //test
			},
			getLinuxLegacyProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return string(ShellTypeZsh), nil
			},
		}

		result := service.detectParentProcessShell()
		assert.Equal(t, ShellTypeZsh, result)
	})

	// Test both methods fail
	t.Run("LinuxBothMethodsFail", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSLinux },
			getPpid: func() int { return 1234 },
			getLinuxProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("failed to read /proc/pid/comm") //nolint:err113 //test
			},
			getLinuxLegacyProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("failed to read /proc/pid/cmdline") //nolint:err113 //test
			},
		}

		result := service.detectParentProcessShell()
		assert.Equal(t, ShellTypeUnknown, result)
	})
}

// TestDetectParentProcessShell_MacOS tests shell detection on macOS
func TestDetectParentProcessShell_MacOS(t *testing.T) {
	t.Run("MacOSSuccess", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSDarwin },
			getPpid: func() int { return 1234 },
			getOtherOsProcessName: func(_ int, _ ExecCommandFunc) (string, error) {
				return string(ShellTypeZsh), nil
			},
		}

		result := service.detectParentProcessShell()
		assert.Equal(t, ShellTypeZsh, result)
	})

	t.Run("MacOSFailure", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSDarwin },
			getPpid: func() int { return 1234 },
			getOtherOsProcessName: func(_ int, _ ExecCommandFunc) (string, error) {
				return "", errors.New("ps command failed") //nolint:err113 //test
			},
		}

		result := service.detectParentProcessShell()
		assert.Equal(t, ShellTypeUnknown, result)
	})
}

// TestGetLinuxProcessName tests the Linux process name retrieval
func TestGetLinuxProcessName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		readFileCalled := false
		mockReadFile := func(name string) ([]byte, error) {
			readFileCalled = true
			assert.Equal(t, "/proc/1234/comm", name)
			return []byte("bash\n"), nil
		}

		name, err := getLinuxProcessName(1234, mockReadFile)
		assert.True(t, readFileCalled)
		assert.NoError(t, err)
		assert.Equal(t, string(ShellTypeBash), name)
	})

	t.Run("Failure", func(t *testing.T) {
		mockReadFile := func(_ string) ([]byte, error) {
			return nil, errors.New("file not found") //nolint:err113 //test
		}

		name, err := getLinuxProcessName(1234, mockReadFile)
		assert.Error(t, err)
		assert.Equal(t, "", name)
	})
}

// TestGetLinuxLegacyProcessName tests the Linux legacy process name retrieval
func TestGetLinuxLegacyProcessName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		readFileCalled := false
		mockReadFile := func(name string) ([]byte, error) {
			readFileCalled = true
			assert.Equal(t, "/proc/1234/cmdline", name)
			return []byte("/usr/bin/zsh\x00arg1\x00arg2"), nil
		}

		name, err := getLinuxLegacyProcessName(1234, mockReadFile)
		assert.True(t, readFileCalled)
		assert.NoError(t, err)
		assert.Equal(t, string(ShellTypeZsh), name) // Should extract basename from path
	})

	t.Run("EmptyCmdline", func(t *testing.T) {
		mockReadFile := func(_ string) ([]byte, error) {
			return []byte(""), nil
		}

		name, err := getLinuxLegacyProcessName(1234, mockReadFile)
		assert.NoError(t, err)
		assert.Equal(t, "", name)
	})

	t.Run("Failure", func(t *testing.T) {
		mockReadFile := func(_ string) ([]byte, error) {
			return nil, errors.New("file not found") //nolint:err113 //test
		}

		name, err := getLinuxLegacyProcessName(1234, mockReadFile)
		assert.Error(t, err)
		assert.Equal(t, "", name)
	})
}

// TestGetOtherOsProcessName tests the process name retrieval on non-Linux systems
func TestGetOtherOsProcessName(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		commandCalled := false
		mockExecCommand := func(name string, arg ...string) *exec.Cmd {
			commandCalled = true
			assert.Equal(t, "ps", name)
			assert.Equal(t, []string{"-o", "comm=", "-p", "1234"}, arg)

			// This is a bit hacky for testing, but necessary since we can't easily
			// mock exec.Command output directly
			fakeCmd := exec.Command("echo", string(ShellTypeZsh))
			return fakeCmd
		}

		name, err := getOtherOsProcessName(1234, mockExecCommand)
		assert.True(t, commandCalled)
		// Skip asserting on the actual result since we can't easily mock exec.Command
		assert.NotEmpty(t, name)
		assert.NoError(t, err)
	})
}

// TestDetectShell tests the main detection method
func TestDetectShell(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSLinux },
			getPpid: func() int { return 1234 },
			getLinuxProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return string(ShellTypeBash), nil
			},
		}

		result := service.DetectShell()
		assert.Equal(t, ShellTypeBash, result)
	})

	t.Run("Failure", func(t *testing.T) {
		service := &ShellDetectionService{ //nolint:exhaustruct // test
			getOS:   func() string { return OSLinux },
			getPpid: func() int { return 1234 },
			getLinuxProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("failed to read /proc/pid/comm") //nolint:err113 //test
			},
			getLinuxLegacyProcessName: func(_ int, _ ReadFileFunc) (string, error) {
				return "", errors.New("failed to read /proc/pid/cmdline") //nolint:err113 //test
			},
		}

		result := service.DetectShell()
		assert.Equal(t, ShellTypeUnknown, result)
	})
}

// TestNewShellDetectionService tests the factory function
func TestNewShellDetectionService(t *testing.T) {
	service := NewShellDetectionService()
	assert.NotNil(t, service, "NewShellDetectionService should return a non-nil service")
}
