package suggest_test

import (
	"math/rand/v2"
	"slices"
	"strings"
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
		{"", []string{"istanbul", "istiyorum", "şehir"}}, // empty prefix: best words overall
		{"iste", []string{"iste", "istemek"}},            // prefix that is itself a word
		{"istem", []string{"istemek"}},                   // ends in the middle of a radix edge
		{"istx", []string{}},                             // diverges in the middle of a radix edge
		{"istanbulx", []string{}},                        // longer than any word
	}
	// One subtest per implementation, so each can be run on its own:
	//   go test ./internal/suggest/ -run 'TestSmall/sortedarray' -v
	for name, s := range implementations(small, k) {
		t.Run(name, func(t *testing.T) {
			for _, tt := range tests {
				got := words(s.Suggest(tt.prefix, k))
				if !slices.Equal(got, tt.want) {
					t.Errorf("Suggest(%q) = %v, want %v", tt.prefix, got, tt.want)
				}
			}
		})
	}
}

// TestK checks k values other than the one the index was built with.
func TestK(t *testing.T) {
	const built = 3
	for name, s := range implementations(small, built) {
		t.Run(name, func(t *testing.T) {
			if got := s.Suggest("is", 0); len(got) != 0 {
				t.Errorf("k=0: got %v, want nothing", words(got))
			}
			// More than the precomputed lists hold: must fall back to a full search.
			want := []string{"istanbul", "istiyorum", "iste", "istemek", "is"}
			if got := words(s.Suggest("is", 20)); !slices.Equal(got, want) {
				t.Errorf("k=20: got %v, want %v", got, want)
			}
		})
	}
}

// bruteForce is the reference answer: scan every word.
func bruteForce(entries []suggest.Entry, prefix string, k int) []suggest.Entry {
	var out []suggest.Entry
	for _, e := range entries {
		if strings.HasPrefix(e.Word, prefix) {
			out = append(out, e)
		}
	}
	suggest.SortEntries(out)
	return out[:min(k, len(out))]
}

// TestAgainstBruteForce compares every implementation to bruteForce on
// generated words. The alphabet is small so words share many prefixes, and it
// includes letters that start with the same UTF-8 byte (ı and ğ are both 0xC4…,
// ç, ö and ü are all 0xC3…) to catch edges split in the middle of a character. Frequencies are drawn from a small range to force ties.
func TestAgainstBruteForce(t *testing.T) {
	alphabet := []rune("aıiğgşsçcöouü")
	rng := rand.New(rand.NewPCG(1, 2))
	seen := map[string]bool{}
	var entries []suggest.Entry
	for len(entries) < 5000 {
		w := make([]rune, 1+rng.IntN(7))
		for i := range w {
			w[i] = alphabet[rng.IntN(len(alphabet))]
		}
		if word := string(w); !seen[word] {
			seen[word] = true
			entries = append(entries, suggest.Entry{Word: word, Freq: uint64(1 + rng.IntN(20))})
		}
	}

	const k = 10
	for name, s := range implementations(entries, k) {
		t.Run(name, func(t *testing.T) {
			for range 3000 {
				w := []rune(entries[rng.IntN(len(entries))].Word)
				prefix := string(w[:rng.IntN(len(w)+1)])
				if rng.IntN(4) == 0 { // sometimes a prefix that may match nothing
					prefix += string(alphabet[rng.IntN(len(alphabet))])
				}
				qk := []int{1, 5, k, 15}[rng.IntN(4)] // 15 > k exercises the precompute fallback
				got, want := s.Suggest(prefix, qk), bruteForce(entries, prefix, qk)
				if !slices.Equal(got, want) {
					t.Fatalf("Suggest(%q, %d) = %v, want %v", prefix, qk, words(got), words(want))
				}
			}
		})
	}
}
