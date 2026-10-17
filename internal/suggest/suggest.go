// Package suggest defines the shared types every autocomplete implementation uses.
package suggest

// Entry is a word and how often it appears in the corpus.
type Entry struct {
	Word string
	Freq uint64
}

// Suggester returns the k most frequent words that start with prefix.
//
// Ordering contract (all implementations must agree, or the shared tests fail):
//   - higher Freq first
//   - equal Freq: Word ascending (byte order)
//
// prefix is assumed to be already normalized (see package normalize).
type Suggester interface {
	Suggest(prefix string, k int) []Entry
}
