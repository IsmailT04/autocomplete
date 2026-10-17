// Package trie implements Suggester with a plain (uncompressed) trie.
package trie

import "autocomplete/internal/suggest"

type node struct {
	// TODO: children — map[rune]*node? map[byte]*node? a sorted slice?
	// Each choice changes memory use a lot; measure it.
	// TODO: word end marker and frequency
	// TODO (precomputed variant): top-k list for this subtree
}

type Trie struct {
	root *node
	// precomputeTopK selects between the two top-k strategies so you can
	// benchmark both with the same code.
	precomputeTopK bool
	k              int
}

// New builds a trie. If precomputeTopK is true, every node stores the best k
// entries in its subtree; otherwise Suggest walks the subtree at query time.
func New(entries []suggest.Entry, k int, precomputeTopK bool) *Trie {
	return &Trie{root: &node{}, precomputeTopK: precomputeTopK, k: k}
}

// Suggest implements suggest.Suggester.
//
// TODO:
//   - walk down to the prefix node
//   - precomputed: return the stored list (what if the caller asks for more than t.k?)
//   - query time: DFS the subtree with a size-k min-heap
func (t *Trie) Suggest(prefix string, k int) []suggest.Entry {
	return nil
}
