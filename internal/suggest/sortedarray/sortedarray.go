// Package sortedarray implements Suggester with a sorted slice and binary search.
package sortedarray

import (
	"container/heap"
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

	// 3. Pick the top k inside the range using a min-heap of size k.
	// Ordering contract:
	// - higher Freq first
	// - equal Freq: Word ascending (so in the min-heap, we want the "smallest" elements at the top
	//   to pop them when capacity exceeds k: lower Freq or (equal Freq and Word descending)).
	h := &entryMinHeap{}
	heap.Init(h)

	for i := start; i < end; i++ {
		entry := ix.entries[i]
		if h.Len() < k {
			heap.Push(h, entry)
		} else {
			// Compare current entry with the minimum element in the heap.
			// If current entry is "better" than the heap minimum, replace it.
			minElem := (*h)[0]
			if compareEntries(entry, minElem) > 0 {
				(*h)[0] = entry
				heap.Fix(h, 0)
			}
		}
	}

	// Extract results from heap in reverse order (highest priority first)
	result := make([]suggest.Entry, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(suggest.Entry)
	}

	return result
}

// compareEntries returns positive if a is strictly better than b according to the contract:
// - higher Freq first
// - equal Freq: Word ascending (lexicographically smaller word is preferred)
func compareEntries(a, b suggest.Entry) int {
	if a.Freq != b.Freq {
		if a.Freq > b.Freq {
			return 1
		}
		return -1
	}
	// Equal Freq: Word ascending means smaller string is better for top-k tie-breaking,
	// but for min-heap eviction, let's make sure heap root is the worst element.
	if a.Word == b.Word {
		return 0
	}
	if a.Word < b.Word {
		return 1 // a is better (comes first alphabetically)
	}
	return -1
}

// entryMinHeap implements container/heap for suggest.Entry where the root is the *worst* element
// (lowest frequency, or highest lexicographical word if frequencies are equal) so it gets evicted.
type entryMinHeap []suggest.Entry

func (h entryMinHeap) Len() int { return len(h) }
func (h entryMinHeap) Less(i, j int) bool {
	if h[i].Freq != h[j].Freq {
		return h[i].Freq < h[j].Freq // lower frequency is smaller (root of min-heap)
	}
	return h[i].Word > h[j].Word // larger word comes last in ascending order, so it's evicted first
}
func (h entryMinHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *entryMinHeap) Push(x interface{}) {
	*h = append(*h, x.(suggest.Entry))
}
func (h *entryMinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
