// Command bench builds every implementation on the same data and reports
// build time, memory, and p50/p99 lookup latency.
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"autocomplete/internal/dataset"
	"autocomplete/internal/suggest"
	"autocomplete/internal/suggest/radix"
	"autocomplete/internal/suggest/sortedarray"
	"autocomplete/internal/suggest/trie"
)

const maxPrefixLen = 5

type impl struct {
	name  string
	build func(es []suggest.Entry, k int) suggest.Suggester
}

var impls = []impl{
	{"sortedarray", func(es []suggest.Entry, k int) suggest.Suggester { return sortedarray.New(es) }},
	{"trie/querytime", func(es []suggest.Entry, k int) suggest.Suggester { return trie.New(es, k, false) }},
	{"trie/precompute", func(es []suggest.Entry, k int) suggest.Suggester { return trie.New(es, k, true) }},
	{"radix/querytime", func(es []suggest.Entry, k int) suggest.Suggester { return radix.New(es, k, false) }},
	{"radix/precompute", func(es []suggest.Entry, k int) suggest.Suggester { return radix.New(es, k, true) }},
}

type query struct {
	prefix string
	runes  int // prefix length in runes (1..maxPrefixLen)
}

type result struct {
	name       string
	build      time.Duration
	heapBytes  uint64
	all        []time.Duration
	byLen      [maxPrefixLen + 1][]time.Duration
	mismatches int
}

func main() {
	path := flag.String("data", "tr_full.txt", "FrequencyWords-style word list")
	size := flag.Int("n", 1_000_000, "number of words to index (top-n by frequency)")
	queries := flag.Int("queries", 100_000, "number of lookups to time")
	k := flag.Int("k", 10, "suggestions per lookup")
	seed := flag.Uint64("seed", 1, "random seed for the query workload")
	only := flag.String("impl", "sortedarray,trie/querytime,trie/precompute",
		"comma-separated implementations to run (add radix/... once it's implemented)")
	csvPath := flag.String("csv", "", "also write results to this CSV file")
	flag.Parse()

	all, err := dataset.Load(*path, 1)
	if err != nil {
		log.Fatal(err)
	}
	entries := dataset.TopN(all, *size)
	all = nil

	var wordBytes int
	for _, e := range entries {
		wordBytes += len(e.Word)
	}
	work := workload(entries, *queries, rand.New(rand.NewPCG(*seed, *seed)))

	fmt.Printf("words=%d  queries=%d  k=%d  seed=%d  %s %s/%s\n",
		len(entries), len(work), *k, *seed, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("word bytes shared by all indexes (not counted below): %.1f MB\n\n", mb(uint64(wordBytes)))

	var results []result
	var reference suggest.Suggester // first implementation; the others are checked against it
	for _, im := range impls {
		if !slices.Contains(strings.Split(*only, ","), im.name) {
			continue
		}
		fmt.Fprintf(os.Stderr, "running %s...\n", im.name)
		r, s := run(im, entries, work, *k, reference)
		if reference == nil {
			reference = s
		}
		results = append(results, r)
	}

	printTable(results, len(entries))
	if *csvPath != "" {
		if err := writeCSV(*csvPath, results, len(entries)); err != nil {
			log.Fatal(err)
		}
	}
}

// workload picks words weighted by frequency (popular words get typed more)
// and cuts each to a random length of 1..maxPrefixLen runes.
func workload(entries []suggest.Entry, n int, rng *rand.Rand) []query {
	cum := make([]uint64, len(entries))
	var total uint64
	for i, e := range entries {
		total += e.Freq
		cum[i] = total
	}
	qs := make([]query, n)
	for i := range qs {
		x := rng.Uint64N(total)
		w := []rune(entries[sort.Search(len(cum), func(j int) bool { return cum[j] > x })].Word)
		l := min(1+rng.IntN(maxPrefixLen), len(w))
		qs[i] = query{prefix: string(w[:l]), runes: l}
	}
	return qs
}

func run(im impl, entries []suggest.Entry, work []query, k int, reference suggest.Suggester) (result, suggest.Suggester) {
	r := result{name: im.name}

	before := heapAlloc()
	start := time.Now()
	s := im.build(entries, k)
	r.build = time.Since(start)
	if after := heapAlloc(); after > before {
		r.heapBytes = after - before
	}

	// Correctness spot check against the first implementation.
	if reference != nil {
		for _, q := range work[:min(2000, len(work))] {
			if !slices.Equal(s.Suggest(q.prefix, k), reference.Suggest(q.prefix, k)) {
				r.mismatches++
			}
		}
	}

	// Warm up caches and the branch predictor before timing.
	for _, q := range work[:min(10_000, len(work))] {
		s.Suggest(q.prefix, k)
	}

	r.all = make([]time.Duration, len(work))
	for i, q := range work {
		t0 := time.Now()
		s.Suggest(q.prefix, k)
		d := time.Since(t0)
		r.all[i] = d
		r.byLen[q.runes] = append(r.byLen[q.runes], d)
	}
	return r, s
}

func heapAlloc() uint64 {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// percentile returns the p-th percentile (0 < p <= 1) using the nearest-rank method.
func percentile(ds []time.Duration, p float64) time.Duration {
	if len(ds) == 0 {
		return 0
	}
	s := slices.Clone(ds)
	slices.Sort(s)
	return s[int(math.Ceil(p*float64(len(s))))-1]
}

func printTable(results []result, words int) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "implementation\tbuild\theap MB\tbytes/word\tp50\tp99\tmax\tmismatches\t")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\t%.1f\t%.0f\t%s\t%s\t%s\t%d\t\n",
			r.name, r.build.Round(time.Millisecond), mb(r.heapBytes), float64(r.heapBytes)/float64(words),
			percentile(r.all, 0.50), percentile(r.all, 0.99), slices.Max(r.all), r.mismatches)
	}
	tw.Flush()

	fmt.Println("\np50 / p99 by prefix length (runes):")
	tw = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprint(tw, "implementation\t")
	for l := 1; l <= maxPrefixLen; l++ {
		fmt.Fprintf(tw, "len %d\t", l)
	}
	fmt.Fprintln(tw)
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t", r.name)
		for l := 1; l <= maxPrefixLen; l++ {
			fmt.Fprintf(tw, "%s / %s\t", percentile(r.byLen[l], 0.50), percentile(r.byLen[l], 0.99))
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func writeCSV(path string, results []result, words int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"implementation", "words", "build_ms", "heap_bytes", "prefix_len", "p50_ns", "p99_ns"})
	for _, r := range results {
		row := func(l string, ds []time.Duration) {
			w.Write([]string{r.name, strconv.Itoa(words), strconv.FormatInt(r.build.Milliseconds(), 10),
				strconv.FormatUint(r.heapBytes, 10), l,
				strconv.FormatInt(percentile(ds, 0.50).Nanoseconds(), 10),
				strconv.FormatInt(percentile(ds, 0.99).Nanoseconds(), 10)})
		}
		row("all", r.all)
		for l := 1; l <= maxPrefixLen; l++ {
			row(strconv.Itoa(l), r.byLen[l])
		}
	}
	w.Flush()
	return w.Error()
}

func mb(b uint64) float64 { return float64(b) / (1 << 20) }
