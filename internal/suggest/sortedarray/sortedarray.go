// Package sortedarray implements Suggester with a sorted slice and binary search.
package sortedarray

import "autocomplete/internal/suggest"

type Index struct {
	// TODO: what do you store? One slice of Entry, or separate slices for
	// words and frequencies? Which is smaller in memory?
}

// New builds the index. It must not modify entries.
func New(entries []suggest.Entry) *Index {
	return &Index{}
}

// Suggest implements suggest.Suggester.
//
// TODO:
//   - find the start of the prefix range (sort.Search / slices.BinarySearchFunc)
//   - find the end of the range: how?
//   - pick the top k inside the range. A min-heap of size k (container/heap)?
func (ix *Index) Suggest(prefix string, k int) []suggest.Entry {
	return nil
}
