// Package trie implements Suggester with a plain (uncompressed) trie.
package trie

import (
	"cmp"
	"slices"

	"autocomplete/internal/suggest"
)

// edge links a node to one child. Children are kept in a slice sorted by r
// instead of a map: most nodes have 1-3 children, and a Go map costs far more
// per node than a short slice (see README, "Children: map vs. sorted slice").
type edge struct {
	r     rune
	child *node
}

type node struct {
	children []edge // sorted by r
	isWord   bool
	entry    suggest.Entry
	topK     []suggest.Entry // precomputed variant only; nil on leaves
}

// child returns the child reached by r, or nil.
func (n *node) child(r rune) *node {
	i, found := slices.BinarySearchFunc(n.children, r, compareEdge)
	if !found {
		return nil
	}
	return n.children[i].child
}

// childOrNew returns the child reached by r, creating it if needed.
func (n *node) childOrNew(r rune) *node {
	i, found := slices.BinarySearchFunc(n.children, r, compareEdge)
	if !found {
		n.children = slices.Insert(n.children, i, edge{r: r, child: &node{}})
	}
	return n.children[i].child
}

func compareEdge(e edge, r rune) int { return cmp.Compare(e.r, r) }

type Trie struct {
	root           *node
	precomputeTopK bool
	k              int
}

// New builds a trie. If precomputeTopK is true, every node stores the best k
// entries in its subtree; otherwise Suggest walks the subtree at query time.
func New(entries []suggest.Entry, k int, precomputeTopK bool) *Trie {
	root := &node{}

	// 1. Insert all entries into the trie
	for _, entry := range entries {
		curr := root
		for _, r := range entry.Word {
			curr = curr.childOrNew(r)
		}
		curr.isWord = true
		curr.entry = entry
	}

	// 2. Precompute top-k for all nodes if requested
	if precomputeTopK && k > 0 {
		precomputeSubtree(root, k)
	}

	return &Trie{
		root:           root,
		precomputeTopK: precomputeTopK,
		k:              k,
	}
}

// precomputeSubtree recursively computes the top-k entries for each node's subtree.
func precomputeSubtree(n *node, k int) []suggest.Entry {
	// A leaf's list would only hold its own word, so don't store one;
	// Suggest answers leaves from n.entry instead.
	if len(n.children) == 0 {
		if n.isWord {
			return []suggest.Entry{n.entry}
		}
		return nil
	}

	var allEntries []suggest.Entry

	if n.isWord {
		allEntries = append(allEntries, n.entry)
	}

	for _, e := range n.children {
		allEntries = append(allEntries, precomputeSubtree(e.child, k)...)
	}

	// Sort and keep top-k for this node. Clone so the stored list doesn't keep
	// the larger backing array (all children's lists combined) alive.
	suggest.SortEntries(allEntries)
	if len(allEntries) > k {
		allEntries = allEntries[:k]
	}
	n.topK = slices.Clone(allEntries)

	return n.topK
}

// Suggest implements suggest.Suggester.
func (t *Trie) Suggest(prefix string, k int) []suggest.Entry {
	if k <= 0 || t.root == nil {
		return nil
	}

	// 1. Walk down to the prefix node
	curr := t.root
	for _, r := range prefix {
		if curr = curr.child(r); curr == nil {
			return nil // Prefix not found
		}
	}

	// 2. Precomputed: return the stored list. If the caller asks for more than
	// t.k, fall back to the query-time DFS.
	if t.precomputeTopK && k <= t.k {
		if len(curr.children) == 0 {
			if curr.isWord {
				return []suggest.Entry{curr.entry}
			}
			return nil
		}
		return slices.Clone(curr.topK[:min(k, len(curr.topK))])
	}

	// 3. Query time: DFS the subtree with a size-k min-heap
	top := suggest.NewTopK(k)
	var dfs func(n *node)
	dfs = func(n *node) {
		if n.isWord {
			top.Offer(n.entry)
		}
		for _, e := range n.children {
			dfs(e.child)
		}
	}
	dfs(curr)

	return top.Result()
}
