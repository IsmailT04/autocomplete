package suggest

import (
	"cmp"
	"container/heap"
	"slices"
)

// Compare orders entries by the Suggester contract: it returns a negative
// number when a ranks before b (higher Freq first, then Word ascending).
// It has the same shape as the func slices.SortFunc expects.
func Compare(a, b Entry) int {
	if a.Freq != b.Freq {
		if a.Freq > b.Freq {
			return -1
		}
		return 1
	}
	return cmp.Compare(a.Word, b.Word)
}

// SortEntries sorts es best first.
func SortEntries(es []Entry) {
	slices.SortFunc(es, Compare)
}

// TopK keeps the k best entries offered to it, using a min-heap of size k
// whose root is the worst entry kept so far.
type TopK struct {
	k int
	h minHeap
}

func NewTopK(k int) *TopK {
	return &TopK{k: k, h: make(minHeap, 0, k)}
}

// Offer considers e for the top k.
func (t *TopK) Offer(e Entry) {
	if t.k <= 0 {
		return
	}
	if len(t.h) < t.k {
		heap.Push(&t.h, e)
		return
	}
	// Replace the worst kept entry only if e ranks before it.
	if Compare(e, t.h[0]) < 0 {
		t.h[0] = e
		heap.Fix(&t.h, 0)
	}
}

// Result returns the kept entries best first. The TopK is empty afterwards.
func (t *TopK) Result() []Entry {
	out := make([]Entry, len(t.h))
	for i := len(out) - 1; i >= 0; i-- {
		out[i] = heap.Pop(&t.h).(Entry)
	}
	return out
}

// minHeap implements container/heap with the worst entry at the root.
type minHeap []Entry

func (h minHeap) Len() int           { return len(h) }
func (h minHeap) Less(i, j int) bool { return Compare(h[i], h[j]) > 0 }
func (h minHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x any)        { *h = append(*h, x.(Entry)) }
func (h *minHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
