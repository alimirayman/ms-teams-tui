package main

import (
	_ "embed"
	"strings"
)

// version is set by release builds with -X main.version=vX.Y.Z.
var version string

//go:embed VERSION
var embeddedVersion string

func displayVersion() string {
	value := strings.TrimSpace(version)
	if value == "" {
		value = strings.TrimSpace(embeddedVersion)
	}
	if !strings.HasPrefix(value, "v") {
		value = "v" + value
	}
	return value
}
