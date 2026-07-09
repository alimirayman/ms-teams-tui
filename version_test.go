package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestVersionFileAndDisplayVersion(t *testing.T) {
	data, err := os.ReadFile("VERSION")
	if err != nil {
		t.Fatal(err)
	}
	value := strings.TrimSpace(string(data))
	if !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(value) {
		t.Fatalf("VERSION must contain semantic version X.Y.Z, got %q", value)
	}

	original := version
	t.Cleanup(func() { version = original })
	version = ""
	if got := displayVersion(); got != "v"+value {
		t.Fatalf("displayVersion() = %q, want %q", got, "v"+value)
	}
	version = "v9.8.7"
	if got := displayVersion(); got != "v9.8.7" {
		t.Fatalf("linked displayVersion() = %q", got)
	}
}
