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

func isCardAttachment(att MessageAttachment) bool {
	if att.ContentType == nil || att.Content == nil || strings.TrimSpace(*att.Content) == "" {
		return false
	}
	contentType := strings.ToLower(strings.TrimSpace(*att.ContentType))
	return strings.HasPrefix(contentType, "application/vnd.microsoft.card.") ||
		strings.HasPrefix(contentType, "application/vnd.microsoft.teams.card.")
}

func messageHasCard(message Message) bool {
	for _, attachment := range message.Attachments {
		if isCardAttachment(attachment) {
			return true
		}
	}
	return false
}

func renderCardAttachment(att MessageAttachment) string {
	if !isCardAttachment(att) {
		return ""
	}
	if isAdaptiveCardAttachment(att) {
		return renderAdaptiveCardContent(*att.Content)
	}
	return renderLegacyCardContent(*att.Content)
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

func renderLegacyCardContent(content string) string {
	var card map[string]any
	if err := json.Unmarshal([]byte(content), &card); err != nil {
		return ""
	}

	var lines []string
	renderLegacyCardObject(card, &lines, true)
	return strings.TrimSpace(strings.Join(compactCardLines(lines), "\n"))
}

func renderLegacyCardObject(card map[string]any, lines *[]string, root bool) {
	appendCardText(lines, stringValue(card["title"]), lipgloss.NewStyle().Bold(true).Foreground(uiTheme.Text))
	appendCardText(lines, stringValue(card["subtitle"]), lipgloss.NewStyle().Foreground(uiTheme.Muted))
	appendCardText(lines, stringValue(card["activityTitle"]), lipgloss.NewStyle().Bold(true))
	appendCardText(lines, stringValue(card["activitySubtitle"]), lipgloss.NewStyle().Foreground(uiTheme.Muted))
	appendCardText(lines, stringValue(card["activityText"]), lipgloss.NewStyle())
	appendCardText(lines, stringValue(card["text"]), lipgloss.NewStyle())

	if root && len(*lines) == 0 {
		appendCardText(lines, stringValue(card["summary"]), lipgloss.NewStyle())
	}
	renderLegacyFacts(anySlice(card["facts"]), lines)

	if code := strings.TrimSpace(stringValue(card["code"])); code != "" {
		if language := strings.TrimSpace(stringValue(card["language"])); language != "" {
			*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Muted).Render(language))
		}
		for _, line := range strings.Split(strings.ReplaceAll(code, "\r\n", "\n"), "\n") {
			*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Text).Render("  "+line))
		}
	}

	for _, item := range anySlice(card["items"]) {
		if child, ok := item.(map[string]any); ok {
			renderLegacyCardObject(child, lines, false)
			appendReceiptPrice(child, lines)
		}
	}
	for _, section := range anySlice(card["sections"]) {
		if child, ok := section.(map[string]any); ok {
			if len(*lines) > 0 {
				*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Border).Render("────────────────────────"))
			}
			renderLegacyCardObject(child, lines, false)
		}
	}

	for _, image := range anySlice(card["images"]) {
		if imageMap, ok := image.(map[string]any); ok {
			label := strings.TrimSpace(stringValue(imageMap["alt"]))
			if label == "" {
				label = strings.TrimSpace(stringValue(imageMap["title"]))
			}
			if label != "" {
				*lines = append(*lines, "🖼️  "+label)
			}
		}
	}

	renderLegacyActions(anySlice(card["buttons"]), lines)
	renderLegacyActions(anySlice(card["actions"]), lines)
	renderLegacyActions(anySlice(card["potentialAction"]), lines)
	if tap, ok := card["tap"].(map[string]any); ok {
		renderLegacyActions([]any{tap}, lines)
	}
}

func appendCardText(lines *[]string, value string, style lipgloss.Style) {
	text := renderCardText(value)
	if text != "" {
		*lines = append(*lines, style.Render(text))
	}
}

func renderCardText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.Contains(text, "<") && strings.Contains(text, ">") {
		return strings.TrimSpace(HTMLToText(text, nil, nil))
	}
	return renderAdaptiveMarkdown(text)
}

func renderLegacyFacts(facts []any, lines *[]string) {
	maxTitle := 0
	for _, raw := range facts {
		if fact, ok := raw.(map[string]any); ok {
			title := legacyFactTitle(fact)
			maxTitle = max(maxTitle, min(24, cellWidth(title)))
		}
	}
	for _, raw := range facts {
		fact, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := legacyFactTitle(fact)
		value := renderCardText(stringValue(fact["value"]))
		if title == "" && value == "" {
			continue
		}
		if value == "" {
			value = "-"
		}
		styledTitle := lipgloss.NewStyle().Bold(true).Foreground(uiTheme.Text).Render(padRight(title, maxTitle))
		*lines = append(*lines, styledTitle+"  "+value)
	}
}

func legacyFactTitle(fact map[string]any) string {
	title := strings.TrimSpace(stringValue(fact["title"]))
	if title == "" {
		title = strings.TrimSpace(stringValue(fact["name"]))
	}
	return title
}

func appendReceiptPrice(item map[string]any, lines *[]string) {
	price := strings.TrimSpace(stringValue(item["price"]))
	quantity := strings.TrimSpace(stringValue(item["quantity"]))
	if price == "" && quantity == "" {
		return
	}
	value := price
	if quantity != "" {
		value = quantity + " × " + price
	}
	*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Muted).Render(value))
}

func renderLegacyActions(actions []any, lines *[]string) {
	for _, raw := range actions {
		action, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title := strings.TrimSpace(stringValue(action["title"]))
		if title == "" {
			title = strings.TrimSpace(stringValue(action["name"]))
		}
		if title == "" {
			title = "Open"
		}
		target := firstCardActionURL(action)
		if target != "" {
			*lines = append(*lines, renderAdaptiveMarkdown(fmt.Sprintf("[%s](%s)", title, target)))
		} else {
			*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Accent).Render("["+title+"]"))
		}
	}
}

func firstCardActionURL(action map[string]any) string {
	for _, key := range []string{"url", "uri", "value"} {
		value := strings.TrimSpace(stringValue(action[key]))
		if strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://") {
			return value
		}
	}
	for _, target := range anySlice(action["targets"]) {
		if targetMap, ok := target.(map[string]any); ok {
			value := strings.TrimSpace(stringValue(targetMap["uri"]))
			if strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://") {
				return value
			}
		}
	}
	return ""
}

func renderAdaptiveElements(elements []any, lines *[]string) {
	for _, raw := range elements {
		element, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		typeName := strings.ToLower(stringValue(element["type"]))
		if boolValue(element["separator"]) && len(*lines) > 0 {
			*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Border).Render("────────────────────────"))
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
				style = style.Bold(true).Foreground(uiTheme.Text)
			}
			if boolValue(element["isSubtle"]) || weight == "lighter" {
				style = style.Foreground(uiTheme.Muted)
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
				title = lipgloss.NewStyle().Bold(true).Foreground(uiTheme.Text).Render(padRight(title, maxTitle))
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
		*lines = append(*lines, lipgloss.NewStyle().Foreground(uiTheme.Accent).Render("["+title+"]"))
	}
}

func cardURLs(content string) []string {
	var value any
	if json.Unmarshal([]byte(content), &value) != nil {
		return nil
	}
	var urls []string
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		case string:
			urls = append(urls, ExtractURLs(markdownToHTML(typed))...)
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
		if isCardAttachment(attachment) {
			urls = append(urls, cardURLs(*attachment.Content)...)
		}
	}
	return dedupeStrings(urls)
}

func cardSignature(attachments []MessageAttachment) string {
	var signature strings.Builder
	for _, att := range attachments {
		if !isCardAttachment(att) {
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
		styledDiagnostic := lipgloss.NewStyle().Foreground(uiTheme.Muted).Render(diagnostic)
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
