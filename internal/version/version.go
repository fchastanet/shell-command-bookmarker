package version

import "runtime/debug"

// VersionValue is the default version when not set by build flags
const VersionValue = "unknown"

// version holds the application version, set at build time or from build info
var version = VersionValue

// Get returns the current application version
func Get() string {
	return version
}

// A user may install this tool using
// `go install github.com/fchastanet/shell-command-bookmarker@latest`
// without -ldflags, in which case the version above is unset. As
// a workaround we use the embedded build version that *is* set when using
// `go install` (and is only set for `go install` and not for `go build`).
func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// < go v1.18
		return
	}
	mainVersion := info.Main.Version
	if mainVersion == "" || mainVersion == "(devel)" {
		// bin not built using `go install`
		return
	}
	// bin built using `go install`
	version = mainVersion
}
