// Command repl is an interactive prompt for trying suggestions by hand.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"autocomplete/internal/dataset"
	"autocomplete/internal/normalize"
	"autocomplete/internal/suggest"
	"autocomplete/internal/suggest/radix"
	"autocomplete/internal/suggest/sortedarray"
	"autocomplete/internal/suggest/trie"
)

func main() {
	path := flag.String("data", "tr_full.txt", "FrequencyWords-style word list")
	impl := flag.String("impl", "radix", "sortedarray | trie | radix")
	size := flag.Int("n", 1_000_000, "number of words to index (top-n by frequency)")
	k := flag.Int("k", 10, "suggestions per lookup")
	precompute := flag.Bool("precompute", true, "trie/radix: store top-k on every node")
	flag.Parse()

	all, err := dataset.Load(*path, 1)
	if err != nil {
		log.Fatal(err)
	}
	entries := dataset.TopN(all, *size)
	all = nil

	start := time.Now()
	var s suggest.Suggester
	switch *impl {
	case "sortedarray":
		s = sortedarray.New(entries)
	case "trie":
		s = trie.New(entries, *k, *precompute)
	case "radix":
		s = radix.New(entries, *k, *precompute)
	default:
		log.Fatalf("unknown -impl %q", *impl)
	}
	fmt.Printf("%s: %d words indexed in %s. Type a prefix (Ctrl-D to quit).\n",
		*impl, len(entries), time.Since(start).Round(time.Millisecond))

	in := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); in.Scan(); fmt.Print("> ") {
		prefix := normalize.Word(in.Text())
		t0 := time.Now()
		res := s.Suggest(prefix, *k)
		d := time.Since(t0)
		for i, e := range res {
			fmt.Printf("%3d. %-24s %d\n", i+1, e.Word, e.Freq)
		}
		fmt.Printf("(%d results in %s)\n", len(res), d)
	}
	fmt.Println()
	if err := in.Err(); err != nil {
		log.Fatal(err)
	}
}
