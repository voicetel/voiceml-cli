// Package main is the VoiceML CLI entry point.
package main

// Version is the CLI's semantic version. Declared as `var` (not `const`) so
// the Makefile can override it at link time via `-ldflags "-X main.Version=…"`.
//
// BuildTime + GitCommit are populated by the Makefile from `date -u` and
// `git rev-parse --short HEAD`. They default to "unknown" so `go build .` (no
// Makefile) still produces a runnable binary.
var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)
