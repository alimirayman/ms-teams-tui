package main

import (
	"strings"
	"testing"
	"time"
)

func TestGetValidTokenSilentFallsBackToAuthenticatedSession(t *testing.T) {
	withTempConfigHome(t)
	clearSessionToken()
	t.Cleanup(clearSessionToken)

	rememberSessionToken(&TokenResponse{
		AccessToken: "session-access-token",
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	})
	got, err := GetValidTokenSilent("client-id")
	if err != nil {
		t.Fatal(err)
	}
	if got != "session-access-token" {
		t.Fatalf("token = %q", got)
	}
}

func TestGetValidTokenSilentRejectsExpiredSessionFallback(t *testing.T) {
	withTempConfigHome(t)
	clearSessionToken()
	t.Cleanup(clearSessionToken)

	rememberSessionToken(&TokenResponse{
		AccessToken: "expired-session-token",
		ExpiresAt:   time.Now().Add(-time.Minute).Unix(),
	})
	_, err := GetValidTokenSilent("client-id")
	if err == nil || !strings.Contains(err.Error(), "no cached token") {
		t.Fatalf("error = %v, want no cached token", err)
	}
}
