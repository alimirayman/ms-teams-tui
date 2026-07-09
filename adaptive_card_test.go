package main

import (
	"strings"
	"testing"
)

const workflowAdaptiveCard = `{
  "type": "AdaptiveCard",
  "body": [
    {
      "type": "TextBlock",
      "size": "large",
      "weight": "bolder",
      "text": "Md Amir Hossain just paid for Premium'25 plan",
      "wrap": true
    },
    {
      "type": "TextBlock",
      "isSubtle": true,
      "text": "User Md Amir Hossain just paid for Premium'25 plan",
      "wrap": true
    },
    {
      "type": "FactSet",
      "facts": [
        {"title": "User", "value": "Md Amir Hossain"},
        {"title": "Phone", "value": "01738587746"},
        {"title": "Amount", "value": ""},
        {"title": "Paid", "value": "Yes"}
      ]
    },
    {
      "type": "Container",
      "separator": true,
      "items": [
        {
          "type": "TextBlock",
          "isSubtle": true,
          "text": "Mir Ayman Ali used a Workflow template. [Get template](https://teams.microsoft.com/l/task/example)",
          "wrap": true
        }
      ]
    }
  ],
  "version": "1.2"
}`

func TestAdaptiveCardMessageRendersStructuredContent(t *testing.T) {
	contentType := "application/vnd.microsoft.card.adaptive"
	body := `<attachment id="card-id"></attachment>`
	appName := "Workflows"
	message := Message{
		ID:   "message-id",
		Body: &MessageBody{Content: &body},
		From: &MessageFrom{Application: &MessageUser{DisplayName: &appName}},
		Attachments: []MessageAttachment{{
			ID:          "card-id",
			ContentType: &contentType,
			Content:     testStringPtr(workflowAdaptiveCard),
		}},
	}

	FilterMessageAttachments(&message)
	if len(message.Attachments) != 1 {
		t.Fatalf("adaptive card was filtered out: %#v", message.Attachments)
	}
	if got := messageSenderName(message); got != "Workflows" {
		t.Fatalf("application sender = %q, want Workflows", got)
	}

	plain := stripANSI(message.GetPlainText())
	for _, expected := range []string{
		"Md Amir Hossain just paid for Premium'25 plan",
		"User    Md Amir Hossain",
		"Phone   01738587746",
		"Amount  -",
		"Paid    Yes",
		"Get template",
	} {
		if !strings.Contains(plain, expected) {
			t.Fatalf("adaptive card output missing %q:\n%s", expected, plain)
		}
	}
	if strings.Contains(plain, "Attachment") {
		t.Fatalf("adaptive card fell back to a generic attachment label:\n%s", plain)
	}
	if strings.Contains(plain, "(https://teams.microsoft.com") {
		t.Fatalf("adaptive card rendered a long link target inline:\n%s", plain)
	}
	if viewable := viewableAttachments(message); len(viewable) != 0 {
		t.Fatalf("adaptive card exposed as downloadable attachment: %#v", viewable)
	}
}

func TestAdaptiveCardURLs(t *testing.T) {
	contentType := "application/vnd.microsoft.card.adaptive"
	urls := messageURLs(Message{Attachments: []MessageAttachment{{
		ContentType: &contentType,
		Content:     testStringPtr(workflowAdaptiveCard),
	}}})
	if len(urls) != 1 || urls[0] != "https://teams.microsoft.com/l/task/example" {
		t.Fatalf("adaptive card URLs = %#v", urls)
	}
}

func TestNonAdaptiveRichCardIsStillFiltered(t *testing.T) {
	contentType := "application/vnd.microsoft.card.hero"
	message := Message{Attachments: []MessageAttachment{{
		ID:          "hero-id",
		ContentType: &contentType,
		Content:     testStringPtr(`{"title":"preview"}`),
	}}}
	FilterMessageAttachments(&message)
	if len(message.Attachments) != 0 {
		t.Fatalf("hero preview card should remain filtered: %#v", message.Attachments)
	}
}
