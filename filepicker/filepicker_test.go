package filepicker

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func loadTestPicker(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"quarterly-report.pdf", "project-plan.docx", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "reports-archive"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("KEY=value"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".config"), 0o700); err != nil {
		t.Fatal(err)
	}

	picker := New()
	picker.CurrentDirectory = dir
	picker.SetHeight(5)
	msg := picker.Init()()
	picker, _ = picker.Update(msg)
	return picker
}

func TestHiddenFilesAndDirectoriesAreVisibleByDefault(t *testing.T) {
	picker := loadTestPicker(t)
	names := entryNames(picker.files)
	if !containsName(names, ".env") || !containsName(names, ".config") {
		t.Fatalf("hidden entries missing from %v", names)
	}
}

func TestTypeAheadFuzzyFilterIsPrimary(t *testing.T) {
	picker := loadTestPicker(t)
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("qtr")})

	if picker.Query != "qtr" {
		t.Fatalf("query = %q, want qtr", picker.Query)
	}
	if len(picker.files) != 1 || picker.files[0].Name() != "quarterly-report.pdf" {
		t.Fatalf("filtered entries = %v", entryNames(picker.files))
	}
}

func TestPrintableNavigationLettersAreSearchInput(t *testing.T) {
	picker := loadTestPicker(t)
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if picker.Query != "j" {
		t.Fatalf("query = %q, want printable j", picker.Query)
	}
}

func TestArrowNavigationAndEnterSelection(t *testing.T) {
	picker := loadTestPicker(t)
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("docx")})
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyEnter})

	selected, path := picker.DidSelectFile(tea.KeyMsg{Type: tea.KeyEnter})
	if !selected {
		t.Fatal("expected filtered file to be selected")
	}
	if filepath.Base(path) != "project-plan.docx" {
		t.Fatalf("selected %q", path)
	}
}

func TestBackspaceEditsQuery(t *testing.T) {
	picker := loadTestPicker(t)
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("notesx")})
	picker, _ = picker.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if picker.Query != "notes" {
		t.Fatalf("query = %q, want notes", picker.Query)
	}
	if len(picker.files) != 1 || picker.files[0].Name() != "notes.txt" {
		t.Fatalf("filtered entries = %v", entryNames(picker.files))
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}

func containsName(names []string, target string) bool {
	for _, name := range names {
		if name == target {
			return true
		}
	}
	return false
}
