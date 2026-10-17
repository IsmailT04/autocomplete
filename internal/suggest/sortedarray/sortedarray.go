// Package sortedarray implements Suggester with a sorted slice and binary search.
package sortedarray

import (
	"sort"
	"strings"

	"autocomplete/internal/suggest"
)

type Index struct {
	// Sorted alphabetically by Word to enable O(log N) prefix range binary search.
	entries []suggest.Entry
}

// New builds the index sorted alphabetically by Word. It does not modify entries.
func New(entries []suggest.Entry) *Index {
	cloned := make([]suggest.Entry, len(entries))
	copy(cloned, entries)

	// Sort alphabetically by Word for binary search range queries.
	sort.Slice(cloned, func(i, j int) bool {
		return cloned[i].Word < cloned[j].Word
	})

	return &Index{
		entries: cloned,
	}
}

// Suggest implements suggest.Suggester.
func (ix *Index) Suggest(prefix string, k int) []suggest.Entry {
	if k <= 0 || len(ix.entries) == 0 {
		return nil
	}

	// 1. Find the start of the prefix range (first index where Word >= prefix)
	start := sort.Search(len(ix.entries), func(i int) bool {
		return ix.entries[i].Word >= prefix
	})

	if start == len(ix.entries) || !strings.HasPrefix(ix.entries[start].Word, prefix) {
		return nil
	}

	// 2. Find the end of the range (first index where Word no longer has prefix).
	// Only search after start: every word there is >= prefix, so "has prefix"
	// is true for a block at the beginning and false after it.
	rest := ix.entries[start:]
	end := start + sort.Search(len(rest), func(i int) bool {
		return !strings.HasPrefix(rest[i].Word, prefix)
	})

	// 3. Pick the top k inside the range with a size-k min-heap.
	top := suggest.NewTopK(k)
	for _, e := range ix.entries[start:end] {
		top.Offer(e)
	}
	return top.Result()
}
