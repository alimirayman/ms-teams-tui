package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// uiTheme centralizes semantic colors so the interface remains coherent on
// both light and dark terminals. The terminal background remains user-owned;
// surfaces are only used for focused or selected elements.
var uiTheme = struct {
	Brand         lipgloss.TerminalColor
	Accent        lipgloss.TerminalColor
	Text          lipgloss.TerminalColor
	Muted         lipgloss.TerminalColor
	Faint         lipgloss.TerminalColor
	Border        lipgloss.TerminalColor
	Surface       lipgloss.TerminalColor
	SurfaceStrong lipgloss.TerminalColor
	Success       lipgloss.TerminalColor
	Warning       lipgloss.TerminalColor
	Danger        lipgloss.TerminalColor
	OwnMessage    lipgloss.TerminalColor
	OtherMessage  lipgloss.TerminalColor
}{
	Brand:         lipgloss.AdaptiveColor{Light: "#464775", Dark: "#A6A7E8"},
	Accent:        lipgloss.AdaptiveColor{Light: "#4F52B2", Dark: "#7B83FF"},
	Text:          lipgloss.AdaptiveColor{Light: "#202124", Dark: "#F2F3F5"},
	Muted:         lipgloss.AdaptiveColor{Light: "#61636B", Dark: "#A3A6B0"},
	Faint:         lipgloss.AdaptiveColor{Light: "#858892", Dark: "#767985"},
	Border:        lipgloss.AdaptiveColor{Light: "#C9CAD3", Dark: "#41434E"},
	Surface:       lipgloss.AdaptiveColor{Light: "#F0F0F7", Dark: "#252630"},
	SurfaceStrong: lipgloss.AdaptiveColor{Light: "#E1E2F2", Dark: "#34364A"},
	Success:       lipgloss.AdaptiveColor{Light: "#0F6B32", Dark: "#62D394"},
	Warning:       lipgloss.AdaptiveColor{Light: "#9A4A00", Dark: "#F2B35F"},
	Danger:        lipgloss.AdaptiveColor{Light: "#B4233C", Dark: "#FF7E93"},
	OwnMessage:    lipgloss.AdaptiveColor{Light: "#23643C", Dark: "#80D6A2"},
	OtherMessage:  lipgloss.AdaptiveColor{Light: "#3D55B4", Dark: "#8EAEFF"},
}

type workspaceMode int

const (
	workspaceNarrow workspaceMode = iota
	workspaceMedium
	workspaceWide
)

type workspaceMetrics struct {
	Mode           workspaceMode
	FrameWidth     int
	SidebarWidth   int
	MessageWidth   int
	SeparatorWidth int
}

func workspaceMetricsForWidth(total int) workspaceMetrics {
	frameWidth := max(1, total-1)
	metrics := workspaceMetrics{FrameWidth: frameWidth}
	switch {
	case total < 72:
		metrics.Mode = workspaceNarrow
		metrics.SidebarWidth = frameWidth
		metrics.MessageWidth = frameWidth
	case total < 110:
		metrics.Mode = workspaceMedium
		metrics.SeparatorWidth = 1
		metrics.SidebarWidth = min(30, max(24, total*30/100))
		metrics.MessageWidth = max(1, frameWidth-metrics.SidebarWidth-metrics.SeparatorWidth)
	default:
		metrics.Mode = workspaceWide
		metrics.SeparatorWidth = 1
		metrics.SidebarWidth = min(38, max(30, total*26/100))
		metrics.MessageWidth = max(1, frameWidth-metrics.SidebarWidth-metrics.SeparatorWidth)
	}
	return metrics
}

func uiRule(width int) string {
	return lipgloss.NewStyle().Foreground(uiTheme.Border).Render(strings.Repeat("─", max(0, width)))
}

func uiBadge(label string, color lipgloss.TerminalColor) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(color).
		Background(uiTheme.Surface).
		Bold(true).
		Padding(0, 1).
		Render(label)
}

func uiKey(key string) string {
	return lipgloss.NewStyle().
		Foreground(uiTheme.Text).
		Background(uiTheme.SurfaceStrong).
		Bold(true).
		Padding(0, 1).
		Render(key)
}

func uiActionHints(actions [][2]string, width int) string {
	parts := make([]string, 0, len(actions))
	for _, action := range actions {
		if action[0] == "" || action[1] == "" {
			continue
		}
		parts = append(parts, uiKey(action[0])+" "+lipgloss.NewStyle().Foreground(uiTheme.Muted).Render(action[1]))
	}
	return fitLine(strings.Join(parts, "  "), width)
}

func uiSectionHeader(label string, count, width int, focused bool) string {
	accent := uiTheme.Muted
	marker := "  "
	if focused {
		accent = uiTheme.Brand
		marker = "▌ "
	}
	left := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(marker + strings.ToUpper(label))
	right := lipgloss.NewStyle().Foreground(uiTheme.Faint).Render(fmt.Sprintf("%d ", count))
	return headerLine(left, right, width)
}

func uiListRow(content string, width int, selected, muted bool) string {
	if width <= 0 {
		return ""
	}
	marker := "  "
	style := lipgloss.NewStyle().Foreground(uiTheme.Text)
	if muted {
		style = style.Foreground(uiTheme.Faint)
	}
	if selected {
		marker = lipgloss.NewStyle().Foreground(uiTheme.Brand).Bold(true).Render("▌ ")
		style = style.Background(uiTheme.SurfaceStrong).Bold(true)
	}
	available := max(0, width-cellWidth(marker))
	return marker + style.Render(padRight(fitLine(content, available), available))
}

func modalDimensions(totalW, totalH, widthPercent, heightPercent, minW, minH, maxW, maxH int) (int, int) {
	w := totalW * widthPercent / 100
	h := totalH * heightPercent / 100
	w = max(minW, w)
	h = max(minH, h)
	if maxW > 0 {
		w = min(w, maxW)
	}
	if maxH > 0 {
		h = min(h, maxH)
	}
	w = min(w, max(1, totalW-2))
	h = min(h, max(1, totalH-2))
	return max(1, w), max(1, h)
}

func renderModalFrame(title, context, body, footer string, width, height int, accent lipgloss.TerminalColor) string {
	contentWidth := max(1, width-4)
	contentHeight := max(1, height-2)

	titleLine := lipgloss.NewStyle().Foreground(accent).Bold(true).Render(title)
	if context != "" {
		titleLine = headerLine(titleLine, lipgloss.NewStyle().Foreground(uiTheme.Muted).Render(context), contentWidth)
	}

	reserved := 1
	if footer != "" {
		reserved += 2
	}
	bodyHeight := max(1, contentHeight-reserved)
	parts := []string{fitLine(titleLine, contentWidth), fitBlock(body, contentWidth, bodyHeight)}
	if footer != "" {
		parts = append(parts, uiRule(contentWidth), fitLine(footer, contentWidth))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(uiTheme.Border).
		Padding(0, 1).
		Width(max(1, width-2)).
		Height(max(1, height-2)).
		Render(strings.Join(parts, "\n"))
}
