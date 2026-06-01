package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Jérémy", "Jeremy"},
		{"François", "Francois"},
		{"jeremy", "jeremy"},
		{"", ""},
		{"München", "Munchen"},
		{"Álvaro", "Alvaro"},
	}

	for _, test := range tests {
		result := normalizeString(test.input)
		if result != test.expected {
			t.Errorf("normalizeString(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestHighlightQuery(t *testing.T) {
	// Force a color profile so that lipgloss outputs ANSI sequences during headless tests.
	lipgloss.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		text     string
		query    string
		contains string // what the highlighted text should contain
	}{
		{"Hello Jérémy", "jeremy", "Jérémy"},
		{"Hello Jeremy", "jeremy", "Jeremy"},
		{"François is here", "francois", "François"},
		{"No match here", "jeremy", "No match here"},
	}

	for _, test := range tests {
		result := highlightQuery(test.text, test.query)
		if test.text == test.contains {
			// If it's a "No match here", result should be identical to input
			if result != test.text {
				t.Errorf("highlightQuery(%q, %q) = %q; expected no match", test.text, test.query, result)
			}
		} else {
			// If it matches, the original substring (with accents) should be in the result,
			// wrapped in some ANSI escape codes.
			if !strings.Contains(result, test.contains) {
				t.Errorf("highlightQuery(%q, %q) = %q; expected to contain %q", test.text, test.query, result, test.contains)
			}
			// And it should not be identical to original text
			if result == test.text {
				t.Errorf("highlightQuery(%q, %q) = %q; expected highlighted output, got original text", test.text, test.query, result)
			}
		}
	}
}
