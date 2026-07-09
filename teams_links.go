package main

import (
	"errors"
	"net/url"
	"strings"
)

func teamsCallURL(chat Chat, withVideo bool) (string, error) {
	var participants []string
	for _, member := range chat.Members {
		if member.Email == nil {
			continue
		}
		if email := strings.TrimSpace(*member.Email); email != "" {
			participants = append(participants, email)
		}
	}
	participants = dedupeStrings(participants)
	if len(participants) == 0 {
		return "", errors.New("selected chat has no callable members")
	}

	query := url.Values{}
	query.Set("users", strings.Join(participants, ","))
	if withVideo {
		query.Set("withVideo", "true")
	}
	return "https://teams.microsoft.com/l/call/0/0?" + query.Encode(), nil
}

func isSelfChatMessages(messages []Message, currentUserID string) bool {
	if currentUserID == "" {
		return false
	}
	foundCurrentUser := false
	for _, message := range messages {
		if message.From == nil {
			continue
		}
		if message.From.Application != nil || message.From.Device != nil {
			return false
		}
		if message.From.User == nil || message.From.User.ID == nil {
			continue
		}
		if *message.From.User.ID != currentUserID {
			return false
		}
		foundCurrentUser = true
	}
	return foundCurrentUser
}
