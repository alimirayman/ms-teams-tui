package main

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// normalizeString decomposes a string, removes Unicode diacritics/accents, and returns it.
// If normalization fails, it returns the input string as-is.
func normalizeString(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return res
}
