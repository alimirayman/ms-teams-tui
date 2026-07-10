package main

import (
	"image"
	"image/color"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func writeClipboardTestPNG(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 30, G: 190, B: 120, A: 255})
		}
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalImagePathsFromPaste(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "clipboard images")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "screenshot.png")
	writeClipboardTestPNG(t, path)

	for _, pasted := range []string{
		path,
		`"` + path + `"`,
		strings.ReplaceAll(path, " ", `\ `),
		(&url.URL{Scheme: "file", Path: path}).String(),
	} {
		got := localImagePathsFromPaste(pasted)
		if len(got) != 1 || got[0] != path {
			t.Fatalf("paste %q resolved to %#v, want %q", pasted, got, path)
		}
	}

	if got := localImagePathsFromPaste("please review " + path); got != nil {
		t.Fatalf("ordinary pasted text was treated as an image: %#v", got)
	}
	if got := localImagePathsFromPaste(filepath.Join(dir, "missing.png")); got != nil {
		t.Fatalf("missing image path was accepted: %#v", got)
	}
}

func TestLocalFilePathsFromPaste(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "dropped files")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "quarterly report.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, pasted := range []string{
		path,
		`"` + path + `"`,
		strings.ReplaceAll(path, " ", `\ `),
		(&url.URL{Scheme: "file", Path: path}).String(),
	} {
		got := localFilePathsFromPaste(pasted)
		if len(got) != 1 || got[0] != path {
			t.Fatalf("paste %q resolved to %#v, want %q", pasted, got, path)
		}
	}
	if got := localFilePathsFromPaste("please review " + path); got != nil {
		t.Fatalf("ordinary pasted text was treated as a file: %#v", got)
	}
}

func TestBracketedImagePathPasteCreatesHostedImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clipboard.png")
	writeClipboardTestPNG(t, path)

	app := NewApp()
	app.InputMode = true
	model := NewModel(app, "client-id", "user-id")
	_ = model.textarea.Focus()
	paste := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(path), Paste: true}

	pending, cmd := model.updateInternal(paste)
	if cmd == nil {
		t.Fatal("expected pasted image path to start an attachment command")
	}
	if pending.pendingFilePastes != 1 {
		t.Fatalf("pending file pastes = %d, want 1", pending.pendingFilePastes)
	}
	if got := pending.textarea.Value(); got != "" {
		t.Fatalf("temporary path leaked into composer: %q", got)
	}

	attached, ok := cmd().(MsgPastedFilesAttached)
	if !ok {
		t.Fatalf("attachment command returned an unexpected message")
	}
	final, _ := pending.updateInternal(attached)
	if got := final.textarea.Value(); got != "[Image 1]" {
		t.Fatalf("composer value = %q, want image placeholder", got)
	}
	if len(final.app.ComposedImages) != 1 {
		t.Fatalf("composed images = %d, want 1", len(final.app.ComposedImages))
	}

	body, _, hosted, _ := formatMessageBodyWithImagesAndFiles(
		final.textarea.Value(),
		nil,
		final.app.ComposedImages,
		nil,
	)
	content, _ := body["content"].(string)
	if len(hosted) != 1 || !strings.Contains(content, `<img src="../hostedContents/1/$value" />`) {
		t.Fatalf("pasted image was not encoded as hosted content: body=%q hosted=%#v", content, hosted)
	}
}

func TestBracketedFilePathPasteCreatesUploadAttachment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "quarterly report.pdf")
	if err := os.WriteFile(path, []byte("%PDF-1.7\nreport"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	app.InputMode = true
	app.Features.FileUpload = true
	model := NewModel(app, "client-id", "user-id")
	_ = model.textarea.Focus()
	paste := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(path), Paste: true}

	pending, cmd := model.updateInternal(paste)
	if cmd == nil {
		t.Fatal("expected pasted file path to start an attachment command")
	}
	if pending.pendingFilePastes != 1 {
		t.Fatalf("pending file pastes = %d, want 1", pending.pendingFilePastes)
	}
	if got := pending.textarea.Value(); got != "" {
		t.Fatalf("temporary path leaked into composer: %q", got)
	}

	attached, ok := cmd().(MsgPastedFilesAttached)
	if !ok {
		t.Fatalf("attachment command returned %T", cmd())
	}
	final, _ := pending.updateInternal(attached)
	if got := final.textarea.Value(); got != "[File: quarterly report.pdf]" {
		t.Fatalf("composer value = %q", got)
	}
	if len(final.app.ComposedFiles) != 1 || final.app.ComposedFiles[0].Name != "quarterly report.pdf" {
		t.Fatalf("composed files = %#v", final.app.ComposedFiles)
	}
}

func TestMixedDroppedPathsAttachImageAndFile(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "screenshot.png")
	writeClipboardTestPNG(t, imagePath)
	filePath := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(filePath, []byte("notes"), 0o600); err != nil {
		t.Fatal(err)
	}

	app := NewApp()
	app.InputMode = true
	app.Features.FileUpload = true
	model := NewModel(app, "client-id", "user-id")
	_ = model.textarea.Focus()
	pasted := imagePath + "\n" + filePath
	pending, cmd := model.updateInternal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(pasted), Paste: true})
	attached, ok := cmd().(MsgPastedFilesAttached)
	if !ok {
		t.Fatalf("attachment command returned an unexpected message")
	}
	final, _ := pending.updateInternal(attached)
	if len(final.app.ComposedImages) != 1 || len(final.app.ComposedFiles) != 1 {
		t.Fatalf("images=%d files=%d", len(final.app.ComposedImages), len(final.app.ComposedFiles))
	}
	if got := final.textarea.Value(); got != "[Image 1][File: notes.txt]" {
		t.Fatalf("composer value = %q", got)
	}
}

func TestBracketedTextPasteRemainsText(t *testing.T) {
	app := NewApp()
	app.InputMode = true
	model := NewModel(app, "client-id", "user-id")
	_ = model.textarea.Focus()
	text := "ordinary pasted text /tmp/not-an-image.txt"

	final, _ := model.updateInternal(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune(text),
		Paste: true,
	})
	if got := final.textarea.Value(); got != text {
		t.Fatalf("ordinary paste = %q, want %q", got, text)
	}
}
