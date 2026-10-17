// Package normalize turns raw words and user input into a single canonical form,
// so "İSTANBUL", "Istanbul" typed on a Turkish keyboard, and "istanbul" all match.
package normalize

import "golang.org/x/text/unicode/norm"

// Word returns the canonical form of s.
//
// TODO:
//   - Turkish-aware lowercasing. Why is strings.ToLower wrong for 'I' and 'İ'?
//     Look at unicode.TurkishCase.
//   - Unicode normalization (NFC). Research golang.org/x/text/unicode/norm.
//     Which step should run first, and does it matter?
//   - Optional: accent folding (ş→s, ğ→g, ...) as a separate function, if you
//     decide "sehir" should match "şehir". Document the decision.
func Word(s string) string {
	return s
}
