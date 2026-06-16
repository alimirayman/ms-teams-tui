package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDBOperations(t *testing.T) {
	// Use cache dir for testing
	cacheDir, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir failed: %v", err)
	}
	dbPath := filepath.Join(cacheDir, "teams-tui-go.db")
	backupPath := dbPath + ".bak"

	// Backup existing DB if any
	dbExists := false
	if _, err := os.Stat(dbPath); err == nil {
		dbExists = true
		_ = os.Rename(dbPath, backupPath)
	}

	defer func() {
		// Restore backup
		CloseDB()
		_ = os.Remove(dbPath)
		if dbExists {
			_ = os.Rename(backupPath, dbPath)
		}
	}()

	// Init
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// Test SaveMessages and GetStoredMessages
	testMessages := []Message{
		{
			ID:              "msg-1",
			CreatedDateTime: "2023-01-01T12:00:00Z",
			MessageType:     "message",
			Body: &MessageBody{
				Content: ptr("Hello World"),
			},
		},
		{
			ID:              "msg-2",
			CreatedDateTime: "2023-01-01T12:05:00Z",
			MessageType:     "message",
			Body: &MessageBody{
				Content: ptr("Second Message"),
			},
		},
	}

	convID := "chat-1"
	if err := SaveMessages(convID, testMessages); err != nil {
		t.Fatalf("SaveMessages failed: %v", err)
	}

	stored, err := GetStoredMessages(convID, 10)
	if err != nil {
		t.Fatalf("GetStoredMessages failed: %v", err)
	}

	if len(stored) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(stored))
	}

	// Newest first check
	if stored[0].ID != "msg-2" {
		t.Errorf("Expected first stored message to be msg-2, got %s", stored[0].ID)
	}

	// Test UpdateStoredMessageBody
	if err := UpdateStoredMessageBody("msg-1", "Hello World (Edited)"); err != nil {
		t.Fatalf("UpdateStoredMessageBody failed: %v", err)
	}

	stored2, err := GetStoredMessages(convID, 10)
	if err != nil {
		t.Fatalf("GetStoredMessages failed: %v", err)
	}

	for _, m := range stored2 {
		if m.ID == "msg-1" {
			if m.Body == nil || m.Body.Content == nil || *m.Body.Content != "Hello World (Edited)" {
				t.Errorf("Expected body to be edited, got %v", m.Body)
			}
		}
	}

	// Test NextLink Save & Get
	testNextLink := "https://graph.microsoft.com/v1.0/chats/chat-1/messages?skipToken=abc"
	if err := SaveNextLink(convID, testNextLink); err != nil {
		t.Fatalf("SaveNextLink failed: %v", err)
	}

	link, err := GetNextLink(convID)
	if err != nil {
		t.Fatalf("GetNextLink failed: %v", err)
	}
	if link != testNextLink {
		t.Errorf("Expected next link to be %s, got %s", testNextLink, link)
	}
}

func ptr[T any](v T) *T {
	return &v
}
