// Package radix implements Suggester with a compressed (radix / Patricia) trie:
// chains of single-child nodes are merged into one edge labeled with a string.
package radix

import (
	"cmp"
	"slices"
	"strings"
	"unicode/utf8"

	"autocomplete/internal/suggest"
)

// edge links a node to one child, keyed by the first rune of the child's label.
// Siblings always differ in their first rune, so the key is unique. (Keying by
// first byte would not be: 'ı' and 'ğ' are both 0xC4 in UTF-8.)
type edge struct {
	r     rune
	child *node
}

type node struct {
	label    string // edge label leading into this node; a substring of a word, so no copy
	children []edge // sorted by r
	isWord   bool
	entry    suggest.Entry
	topK     []suggest.Entry // precomputed variant only; nil on leaves
}

func (n *node) find(r rune) (int, bool) {
	return slices.BinarySearchFunc(n.children, r, func(e edge, r rune) int { return cmp.Compare(e.r, r) })
}

func (n *node) addChild(c *node) {
	r, _ := utf8.DecodeRuneInString(c.label)
	i, _ := n.find(r)
	n.children = slices.Insert(n.children, i, edge{r: r, child: c})
}

type Tree struct {
	root           *node
	precomputeTopK bool
	k              int
}

// New builds the tree. If precomputeTopK is true, every node stores the best k
// entries in its subtree; otherwise Suggest walks the subtree at query time.
func New(entries []suggest.Entry, k int, precomputeTopK bool) *Tree {
	root := &node{}
	for _, e := range entries {
		insert(root, e)
	}
	if precomputeTopK && k > 0 {
		precomputeSubtree(root, k)
	}
	return &Tree{root: root, precomputeTopK: precomputeTopK, k: k}
}

// insert adds e below n. The three cases from drawing it on paper:
//   - no child starts with the next rune: add a leaf with the rest of the word
//   - the word continues past a child's whole label: descend into it
//   - the word ends or diverges in the middle of a label: split the edge
func insert(n *node, e suggest.Entry) {
	rest := e.Word
	for rest != "" {
		r, _ := utf8.DecodeRuneInString(rest)
		i, found := n.find(r)
		if !found {
			n.addChild(&node{label: rest, isWord: true, entry: e})
			return
		}
		c := n.children[i].child
		l := commonPrefix(c.label, rest)
		if l < len(c.label) {
			// Split c: a new middle node takes the shared part of the label,
			// c keeps the remainder and becomes the middle node's only child.
			mid := &node{label: c.label[:l]}
			c.label = c.label[l:]
			mid.addChild(c)
			n.children[i].child = mid
			c = mid
		}
		n, rest = c, rest[l:]
	}
	n.isWord = true
	n.entry = e
}

// commonPrefix returns the length in bytes of the longest common prefix of a
// and b, cut back to a rune boundary so labels are always valid UTF-8.
func commonPrefix(a, b string) int {
	l := 0
	for l < len(a) && l < len(b) && a[l] == b[l] {
		l++
	}
	for l > 0 && l < len(a) && !utf8.RuneStart(a[l]) {
		l--
	}
	return l
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

	// Clone so the stored list doesn't keep the larger backing array alive.
	suggest.SortEntries(allEntries)
	if len(allEntries) > k {
		allEntries = allEntries[:k]
	}
	n.topK = slices.Clone(allEntries)
	return n.topK
}

// Suggest implements suggest.Suggester.
func (t *Tree) Suggest(prefix string, k int) []suggest.Entry {
	if k <= 0 || t.root == nil {
		return nil
	}

	// 1. Walk down to the node whose subtree holds exactly the words with this
	// prefix. The prefix can end in the middle of an edge label; then every
	// word below that edge still matches.
	curr, rest := t.root, prefix
	for rest != "" {
		r, _ := utf8.DecodeRuneInString(rest)
		i, found := curr.find(r)
		if !found {
			return nil
		}
		c := curr.children[i].child
		if len(rest) <= len(c.label) {
			if !strings.HasPrefix(c.label, rest) {
				return nil
			}
			curr = c
			break
		}
		if !strings.HasPrefix(rest, c.label) {
			return nil
		}
		curr, rest = c, rest[len(c.label):]
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
