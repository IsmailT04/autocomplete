# Turkish Autocomplete: Data Structure Comparison

Given a prefix, return the top-k most frequent Turkish words that start with it.
Three index structures × two top-k strategies, benchmarked side by side.

## Dataset

- Source: [hermitdave/FrequencyWords](https://github.com/hermitdave/FrequencyWords) `tr_full.txt`
- Raw size: 2,035,629 lines
- After normalization + dedup: TODO
- Cutoff used: TODO

## Running

```sh
go test ./...
go run ./cmd/bench -data tr_full.txt -n 1000000
go run ./cmd/repl -impl trie
```

## Decisions

### Trie vs. sorted array vs. compressed trie
TODO

### Top-k: precomputed per node vs. computed at query time
TODO

### Turkish characters
TODO (casing: I/ı/İ/i, NFC, accent folding: yes/no and why)

## Results

Apple M5, 16 GB, go1.27.2 darwin/arm64. n = 1,000,000 words, k = 10,
100,000 queries, seed 1. Single run of `go run ./cmd/bench -n 1000000`.

| Implementation | Build time | Memory | Bytes/word | p50 | p99 | p50 / p99 (1-char prefix) |
|---|---|---|---|---|---|---|
| sortedarray | 162 ms | 22.9 MB | 24 | 7.0 µs | 238 µs | 101 µs / 245 µs |
| trie / query-time | 274 ms | 192.7 MB | 202 | 46.3 µs | 1.99 ms | 923 µs / 2.18 ms |
| trie / precomputed | 473 ms | 276.8 MB | 290 | 83 ns | 500 ns | 42 ns / 292 ns |
| radix / query-time | | | | | | |
| radix / precomputed | | | | | | |

### How it's measured (`cmd/bench`)

- **Memory:** `runtime.HeapAlloc` after two GCs, before vs. after building. The word
  strings (10.3 MB) are shared with the input and not counted for any implementation.
- **Workload:** words picked weighted by frequency (popular words get typed more),
  cut to a random length of 1–5 runes. Same queries for every implementation.
- **Latency:** 10,000 warm-up queries, then each query timed on its own with
  `time.Now()`. Includes allocating the result slice. GC pauses can show up in
  p99/max; that's real behavior, not noise to remove.
- **Correctness:** the first 2,000 queries are checked against the first
  implementation run (`mismatches` column).

### Trie experiments

Same bench, 1M words. Each row adds one change to the row above.

| Trie version | query-time memory | precomputed memory | query-time p50 / p99 (1-char) |
|---|---|---|---|
| `map[rune]*node`, map created for every node, top-k slices keep extra capacity | 415.8 MB | 562.9 MB | 4.4 ms / 10.4 ms |
| + map only created on first child, top-k lists cloned to exact size | 384.9 MB | 484.4 MB | 4.4 ms / 10.4 ms |
| + no top-k list stored on leaves | — | 468.9 MB | — |
| + children as a sorted slice of `{rune, *node}` instead of a map | **192.7 MB** | **276.8 MB** | **0.92 ms / 2.18 ms** |

TODO: explain why the sorted slice wins (per-map overhead, how many children most
nodes have, pointer chasing vs. contiguous memory), and why sortedarray still
beats the query-time trie on short prefixes.

## What I learned
TODO
