package suggest_test

import (
	"slices"
	"testing"

	"autocomplete/internal/suggest"
	"autocomplete/internal/suggest/radix"
	"autocomplete/internal/suggest/sortedarray"
	"autocomplete/internal/suggest/trie"
)

// A tiny hand-checkable dataset. Every implementation must return exactly the
// same answers for it.
var small = []suggest.Entry{
	{"istanbul", 900},
	{"istiyorum", 700},
	{"istemek", 300},
	{"iste", 300},
	{"is", 50},
	{"ışık", 400},
	{"ılık", 20},
	{"şehir", 500},
	{"şeker", 200},
}

func implementations(entries []suggest.Entry, k int) map[string]suggest.Suggester {
	return map[string]suggest.Suggester{
		"sortedarray":      sortedarray.New(entries),
		"trie/querytime":   trie.New(entries, k, false),
		"trie/precompute":  trie.New(entries, k, true),
		"radix/querytime":  radix.New(entries, k, false),
		"radix/precompute": radix.New(entries, k, true),
	}
}

func words(es []suggest.Entry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.Word
	}
	return out
}

func TestSmall(t *testing.T) {
	const k = 3
	tests := []struct {
		prefix string
		want   []string
	}{
		{"ist", []string{"istanbul", "istiyorum", "iste"}}, // iste vs istemek: tie → byte order
		{"is", []string{"istanbul", "istiyorum", "iste"}},
		{"ı", []string{"ışık", "ılık"}}, // dotless ı must not match i
		{"şe", []string{"şehir", "şeker"}},
		{"istanbul", []string{"istanbul"}},
		{"x", []string{}},
		// TODO: empty prefix? k larger than matches? k = 0? Decide and add cases.
	}
	for name, s := range implementations(small, k) {
		for _, tt := range tests {
			got := words(s.Suggest(tt.prefix, k))
			if !slices.Equal(got, tt.want) {
				t.Errorf("%s: Suggest(%q) = %v, want %v", name, tt.prefix, got, tt.want)
			}
		}
	}
}

// TODO: TestAgainstBruteForce — load a sample of the real dataset, write a
// dead-simple brute-force Suggest (scan everything), and check every
// implementation agrees with it on many random prefixes.
