// Command bench builds every implementation on the same data and reports
// build time, memory, and p50/p99 lookup latency.
package main

import (
	"flag"
	"fmt"
	"log"

	"autocomplete/internal/dataset"
)

func main() {
	path := flag.String("data", "tr_full.txt", "FrequencyWords-style word list")
	size := flag.Int("n", 1_000_000, "number of words to index (top-n by frequency)")
	queries := flag.Int("queries", 100_000, "number of lookups to time")
	k := flag.Int("k", 10, "suggestions per lookup")
	flag.Parse()

	entries, err := dataset.Load(*path, 1)
	if err != nil {
		log.Fatal(err)
	}
	entries = dataset.TopN(entries, *size)
	fmt.Printf("loaded %d entries, queries=%d, k=%d\n", len(entries), *queries, *k)

	// TODO: generate the query workload once, so every implementation gets the
	// same prefixes. Pick words weighted by Freq, cut to length 1..5 (in runes,
	// not bytes!). Use a fixed rand seed so runs are reproducible.

	// TODO: for each implementation:
	//   build time: time.Now() / time.Since around New(...)
	//   memory:     runtime.GC() + runtime.ReadMemStats before and after building;
	//               HeapAlloc difference. Keep the index alive until after measuring
	//               (runtime.KeepAlive).
	//   latency:    warm up, then time each Suggest call individually, sort the
	//               durations, read p50 and p99. Also split by prefix length.

	// TODO: print a table (text/tabwriter) and maybe write CSV for charts.
}
