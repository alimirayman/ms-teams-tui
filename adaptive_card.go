package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func isAdaptiveCardAttachment(att MessageAttachment) bool {
	return att.ContentType != nil && strings.EqualFold(
		strings.TrimSpace(*att.ContentType),
		"application/vnd.microsoft.card.adaptive",
	)
}

func messageHasAdaptiveCard(message Message) bool {
	for _, attachment := range message.Attachments {
		if isAdaptiveCardAttachment(attachment) {
			return true
		}
	}
	return false
}

func renderAdaptiveCardAttachment(att MessageAttachment) string {
	if !isAdaptiveCardAttachment(att) || att.Content == nil {
		return ""
	}
	return renderAdaptiveCardContent(*att.Content)
}

func renderAdaptiveCardContent(content string) string {
	var card map[string]any
	if err := json.Unmarshal([]byte(content), &card); err != nil {
		return ""
	}

	var lines []string
	renderAdaptiveElements(anySlice(card["body"]), &lines)
	renderAdaptiveActions(anySlice(card["actions"]), &lines)
	return strings.TrimSpace(strings.Join(compactCardLines(lines), "\n"))
}

func renderAdaptiveElements(elements []any, lines *[]string) {
	for _, raw := range elements {
		element, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		typeName := strings.ToLower(stringValue(element["type"]))
		if boolValue(element["separator"]) && len(*lines) > 0 {
			*lines = append(*lines, lipgloss.NewStyle().Foreground(colDimGray).Render("────────────────────────"))
		}

		switch typeName {
		case "textblock":
			text := renderAdaptiveMarkdown(stringValue(element["text"]))
			if text == "" {
				continue
			}
			style := lipgloss.NewStyle()
			weight := strings.ToLower(stringValue(element["weight"]))
			size := strings.ToLower(stringValue(element["size"]))
			if weight == "bolder" || weight == "bold" || size == "large" || size == "extralarge" {
				style = style.Bold(true).Foreground(colWhite)
			}
			if boolValue(element["isSubtle"]) || weight == "lighter" {
				style = style.Foreground(colDimGray)
			}
			*lines = append(*lines, style.Render(text))

		case "richtextblock":
			var parts []string
			for _, inlineRaw := range anySlice(element["inlines"]) {
				if inline, ok := inlineRaw.(map[string]any); ok {
					parts = append(parts, stringValue(inline["text"]))
				}
			}
			if text := renderAdaptiveMarkdown(strings.Join(parts, "")); text != "" {
				*lines = append(*lines, text)
			}

		case "factset":
			facts := anySlice(element["facts"])
			maxTitle := 0
			for _, factRaw := range facts {
				if fact, ok := factRaw.(map[string]any); ok {
					maxTitle = max(maxTitle, min(24, cellWidth(strings.TrimSpace(stringValue(fact["title"])))))
				}
			}
			for _, factRaw := range facts {
				fact, ok := factRaw.(map[string]any)
				if !ok {
					continue
				}
				title := strings.TrimSpace(stringValue(fact["title"]))
				value := renderAdaptiveMarkdown(strings.TrimSpace(stringValue(fact["value"])))
				if value == "" {
					value = "-"
				}
				title = lipgloss.NewStyle().Bold(true).Foreground(colWhite).Render(padRight(title, maxTitle))
				*lines = append(*lines, title+"  "+value)
			}

		case "image":
			label := strings.TrimSpace(stringValue(element["altText"]))
			if label == "" {
				label = "Card image"
			}
			*lines = append(*lines, "🖼️  "+label)

		case "actionset":
			renderAdaptiveActions(anySlice(element["actions"]), lines)

		case "container", "column":
			renderAdaptiveElements(anySlice(element["items"]), lines)

		case "columnset":
			for _, columnRaw := range anySlice(element["columns"]) {
				if column, ok := columnRaw.(map[string]any); ok {
					renderAdaptiveElements(anySlice(column["items"]), lines)
				}
			}

		default:
			renderAdaptiveElements(anySlice(element["items"]), lines)
			renderAdaptiveElements(anySlice(element["body"]), lines)
			renderAdaptiveActions(anySlice(element["actions"]), lines)
		}
	}
}

func renderAdaptiveActions(actions []any, lines *[]string) {
	for _, raw := range actions {
		action, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := strings.TrimSpace(stringValue(action["title"]))
		if title == "" {
			title = "Open"
		}
		if target := strings.TrimSpace(stringValue(action["url"])); target != "" {
			*lines = append(*lines, renderAdaptiveMarkdown(fmt.Sprintf("[%s](%s)", title, target)))
			continue
		}
		*lines = append(*lines, lipgloss.NewStyle().Foreground(colCyan).Render("["+title+"]"))
	}
}

func adaptiveCardURLs(content string) []string {
	var value any
	if json.Unmarshal([]byte(content), &value) != nil {
		return nil
	}
	var urls []string
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			for key, child := range typed {
				if strings.EqualFold(key, "url") {
					if target, ok := child.(string); ok && strings.TrimSpace(target) != "" {
						urls = append(urls, target)
					}
				}
				if strings.EqualFold(key, "text") {
					if text, ok := child.(string); ok {
						urls = append(urls, ExtractURLs(markdownToHTML(text))...)
					}
				}
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return dedupeStrings(urls)
}

func messageURLs(message Message) []string {
	var urls []string
	if message.Body != nil && message.Body.Content != nil {
		urls = append(urls, ExtractURLs(*message.Body.Content)...)
	}
	for _, attachment := range message.Attachments {
		if isAdaptiveCardAttachment(attachment) && attachment.Content != nil {
			urls = append(urls, adaptiveCardURLs(*attachment.Content)...)
		}
	}
	return dedupeStrings(urls)
}

func adaptiveCardSignature(attachments []MessageAttachment) string {
	var signature strings.Builder
	for _, att := range attachments {
		if !isAdaptiveCardAttachment(att) || att.Content == nil {
			continue
		}
		signature.WriteString(att.ID)
		signature.WriteByte('\n')
		signature.WriteString(*att.Content)
		signature.WriteByte('\n')
	}
	return signature.String()
}

func renderAdaptiveMarkdown(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	html := markdownToHTML(text)
	rendered := HTMLToText(html, nil, nil)
	for _, target := range ExtractURLs(html) {
		diagnostic := " (" + target + ")"
		rendered = strings.ReplaceAll(rendered, diagnostic, "")
		styledDiagnostic := lipgloss.NewStyle().Foreground(colDimGray).Render(diagnostic)
		rendered = strings.ReplaceAll(rendered, styledDiagnostic, "")
	}
	return rendered
}

func compactCardLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	previousBlank := true
	for _, line := range lines {
		blank := strings.TrimSpace(stripANSI(line)) == ""
		if blank && previousBlank {
			continue
		}
		result = append(result, line)
		previousBlank = blank
	}
	for len(result) > 0 && strings.TrimSpace(stripANSI(result[len(result)-1])) == "" {
		result = result[:len(result)-1]
	}
	return result
}

func anySlice(value any) []any {
	if values, ok := value.([]any); ok {
		return values
	}
	return nil
}

func stringValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func boolValue(value any) bool {
	if flag, ok := value.(bool); ok {
		return flag
	}
	return false
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
