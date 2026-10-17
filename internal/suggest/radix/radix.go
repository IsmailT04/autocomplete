// Package radix implements Suggester with a compressed (radix / Patricia) trie:
// chains of single-child nodes are merged into one edge labeled with a string.
package radix

import "autocomplete/internal/suggest"

type node struct {
	// TODO: edge label, children, word end + frequency, optional top-k list
}

type Tree struct {
	root           *node
	precomputeTopK bool
	k              int
}

// New builds the tree.
//
// TODO: insertion is the hard part. Draw the cases on paper first:
//   - new word shares no prefix with any edge
//   - new word ends in the middle of an edge (split the edge)
//   - new word diverges in the middle of an edge (split, then add a child)
func New(entries []suggest.Entry, k int, precomputeTopK bool) *Tree {
	return &Tree{root: &node{}, precomputeTopK: precomputeTopK, k: k}
}

// Suggest implements suggest.Suggester.
// Careful: the prefix can end in the middle of an edge label.
func (t *Tree) Suggest(prefix string, k int) []suggest.Entry {
	return nil
}
