// Package buildinfo holds the version metadata stamped into every binary at
// build time. The server, CLI, and MCP server all report from here, so there
// is exactly one set of ldflags to maintain:
//
//	-X github.com/nimbleflux/fluxbase/internal/buildinfo.version=2026.10.1
//	-X github.com/nimbleflux/fluxbase/internal/buildinfo.commit=<sha>
//	-X github.com/nimbleflux/fluxbase/internal/buildinfo.buildDate=<rfc3339>
//
// Unstamped builds (go run, go build, go install) fall back to the module
// info the Go toolchain embeds, so dev builds self-identify instead of
// reporting a bare "dev".
package buildinfo

import "runtime/debug"

// Overridden via -ldflags "-X ..." at build time. Empty means unstamped.
var (
	version   = ""
	commit    = ""
	buildDate = ""
)

// Version returns the release version. Fallback order: ldflags stamp, then
// the embedded module version (the real version for `go install ...@v...`),
// then "dev".
func Version() string {
	if version != "" {
		return version
	}
	if v, ok := moduleVersion(); ok && v != "" && v != "(devel)" {
		return v
	}
	return "dev"
}

// Commit returns the build commit. Falls back to the VCS revision embedded
// by the toolchain, else "unknown".
func Commit() string {
	if commit != "" {
		return commit
	}
	if rev, ok := vcsRevision(); ok && rev != "" {
		return rev
	}
	return "unknown"
}

// BuildDate returns the build timestamp, or "unknown" when unstamped (the
// toolchain does not record one).
func BuildDate() string {
	if buildDate != "" {
		return buildDate
	}
	return "unknown"
}

func moduleVersion() (string, bool) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	return bi.Main.Version, true
}

func vcsRevision() (string, bool) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false
	}
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" {
			return s.Value, true
		}
	}
	return "", false
}
