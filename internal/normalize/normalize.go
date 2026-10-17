// Package normalize turns raw words and user input into a single canonical form,
// so "İSTANBUL", "Istanbul" typed on a Turkish keyboard, and "istanbul" all match.
package normalize

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/norm"
)

var turkishLower = cases.Lower(language.Turkish)

// Word returns the canonical form of s.
func Word(s string) string {
	normalized := norm.NFC.String(s)

	lowercased := turkishLower.String(normalized)

	return lowercased
}
