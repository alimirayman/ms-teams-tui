package main

import (
	"net/url"
	"testing"
)

func TestTeamsCallURL(t *testing.T) {
	first := "first@example.com"
	second := "second@example.com"
	chat := Chat{Members: []ChatMember{{Email: &first}, {Email: &second}}}

	target, err := teamsCallURL(chat, true)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := parsed.Query().Get("users"); got != "first@example.com,second@example.com" {
		t.Fatalf("users = %q", got)
	}
	if got := parsed.Query().Get("withVideo"); got != "true" {
		t.Fatalf("withVideo = %q", got)
	}
}

func TestSelfChatClassification(t *testing.T) {
	currentID := "current-user"
	otherID := "other-user"
	appID := "workflows"

	if !isSelfChatMessages([]Message{{From: &MessageFrom{User: &MessageUser{ID: &currentID}}}}, currentID) {
		t.Fatal("current-user-only conversation was not classified as self chat")
	}
	if isSelfChatMessages([]Message{{From: &MessageFrom{User: &MessageUser{ID: &otherID}}}}, currentID) {
		t.Fatal("other-user conversation was classified as self chat")
	}
	if isSelfChatMessages([]Message{{From: &MessageFrom{Application: &MessageUser{ID: &appID}}}}, currentID) {
		t.Fatal("application conversation was classified as self chat")
	}
}
