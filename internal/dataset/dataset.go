// Package dataset loads a FrequencyWords-style list ("word count" per line).
package dataset

import "autocomplete/internal/suggest"

// Load reads path, normalizes every word, merges duplicates that collapse to the
// same normalized form (summing their counts), and drops words with
// Freq < minFreq.
//
// The returned order is up to you; decide whether callers should sort it themselves.
//
// TODO:
//   - read line by line (bufio.Scanner), split word and count
//   - normalize.Word on each word
//   - merge duplicates with a map
//   - what do you do with malformed lines?
func Load(path string, minFreq uint64) ([]suggest.Entry, error) {
	return nil, nil
}

// TopN returns the n most frequent entries. Useful for benchmarking at exact
// sizes (100k, 500k, 1M, ...).
func TopN(entries []suggest.Entry, n int) []suggest.Entry {
	return nil
}
