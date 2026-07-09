package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSendCmuxNotification(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-specific")
	}

	dir := t.TempDir()
	output := filepath.Join(dir, "args")
	cli := filepath.Join(dir, "cmux")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$CMUX_TEST_OUTPUT\"\n"
	if err := os.WriteFile(cli, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CMUX_BUNDLED_CLI_PATH", cli)
	t.Setenv("CMUX_TEST_OUTPUT", output)

	if !sendCmuxNotification("Workflows", "Payment received") {
		t.Fatal("cmux notification was not sent")
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSpace(string(data)), "\n")
	want := []string{"notify", "--title", "Microsoft Teams", "--subtitle", "Workflows", "--body", "Payment received"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("cmux args = %#v, want %#v", got, want)
	}
}
