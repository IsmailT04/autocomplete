# Turkish Autocomplete: Data Structure Comparison

Given a prefix, return the top-k most frequent Turkish words that start with it.
Three index structures × two top-k strategies, benchmarked side by side on 1M
real Turkish words.

**Short version** (1M words, k = 10, Apple M5):

| | Memory | p99 lookup | When to pick it |
|---|---|---|---|
| Sorted array | **22.9 MB** | 253 µs | Memory matters most (mobile, embedded). Still far below a keystroke's time budget. |
| Radix tree + precomputed top-k | 177.1 MB | **667 ns** | Lookups/sec matter most (a server). |
| Plain trie | 193–277 MB | 0.5 µs–2.2 ms | Never: the radix tree beats it on memory and speed. |

## Dataset

- Source: [hermitdave/FrequencyWords](https://github.com/hermitdave/FrequencyWords) `tr_full.txt`
  (word frequencies from Turkish OpenSubtitles)
- Raw: 2,035,629 lines
- After Turkish lowercasing, NFC and merging duplicates: **2,004,272 unique words**
- Main results use the top 1,000,000 by frequency (`-n 1000000`); the scaling table uses all of them.
- Not in git (28 MB, see `.gitignore`). Download `tr_full.txt` into the repo root.

## Running

```sh
go test ./...                              # all implementations vs. hand-checked + brute-force answers
go run ./cmd/bench -n 1000000              # build time, memory, p50/p99 for every implementation
go run ./cmd/bench -impl sortedarray,radix/precompute -csv results.csv
go run ./cmd/repl                          # type prefixes, see suggestions (-impl, -precompute, -n, -k)
```

## Layout

```
internal/normalize          Turkish lowercasing + NFC (used for both data and user input)
internal/dataset            load "word count" lines, merge duplicates, sort by frequency
internal/suggest            Entry, Suggester interface, ordering rule, shared top-k heap, tests
internal/suggest/sortedarray
internal/suggest/trie       plain trie, one node per rune
internal/suggest/radix      compressed trie, one node per branch point
cmd/bench                   benchmark harness
cmd/repl                    interactive prompt
```

Every implementation returns exactly the same answers: higher frequency first, ties by
word. `go test` checks this against a brute-force scan on 5,000 generated words, and
the bench checks 2,000 real queries against sortedarray (`mismatches` column).

## Results

Apple M5, 16 GB, go1.27.2 darwin/arm64. n = 1,000,000 words, k = 10, 100,000 queries,
seed 1. Single run of `go run ./cmd/bench -n 1000000`; latencies move by ~10–20%
between runs, memory doesn't.

| Implementation | Build time | Memory | Bytes/word | p50 | p99 | max |
|---|---|---|---|---|---|---|
| sortedarray | 166 ms | 22.9 MB | 24 | 8.1 µs | 253 µs | 2.6 ms |
| trie / query-time | 354 ms | 192.7 MB | 202 | 51.5 µs | 2.24 ms | 17.9 ms |
| trie / precomputed | 545 ms | 276.8 MB | 290 | 84 ns | 542 ns | 18.0 µs |
| radix / query-time | 280 ms | 125.4 MB | 131 | 30.0 µs | 1.11 ms | 4.8 ms |
| radix / precomputed | 407 ms | 177.1 MB | 186 | 84 ns | 667 ns | 14.5 µs |

**By prefix length (p50 / p99).** Short prefixes match many words, so they are the hard case:

| Implementation | 1 letter | 2 letters | 3 letters | 5 letters |
|---|---|---|---|---|
| sortedarray | 111 µs / 385 µs | 17.5 µs / 108 µs | 5.5 µs / 21.7 µs | 2.9 µs / 9.6 µs |
| trie / query-time | 1.01 ms / 2.65 ms | 142 µs / 886 µs | 31.5 µs / 128 µs | 7.5 µs / 52.2 µs |
| trie / precomputed | 83 ns / 292 ns | 83 ns / 333 ns | 125 ns / 375 ns | 167 ns / 709 ns |
| radix / query-time | 520 µs / 1.27 ms | 75.1 µs / 464 µs | 18.9 µs / 71.7 µs | 5.5 µs / 31.5 µs |
| radix / precomputed | 83 ns / 375 ns | 83 ns / 416 ns | 125 ns / 458 ns | 167 ns / 792 ns |

**Scaling (memory in bytes per word, build time).** 20,000 queries per run.

| Implementation | 100k | 500k | 1M | 2.0M (all) |
|---|---|---|---|---|
| sortedarray | 24 B, 16 ms | 24 B, 80 ms | 24 B, 166 ms | 24 B, 366 ms |
| trie / query-time | 192 B, 25 ms | 192 B, 150 ms | 202 B, 354 ms | 231 B, 509 ms |
| trie / precomputed | 282 B, 46 ms | 280 B, 242 ms | 290 B, 545 ms | 323 B, 941 ms |
| radix / query-time | 132 B, 21 ms | 131 B, 140 ms | 131 B, 280 ms | 133 B, 493 ms |
| radix / precomputed | 188 B, 35 ms | 186 B, 208 ms | 186 B, 407 ms | 186 B, 765 ms |

### How it's measured (`cmd/bench`)

- **Memory:** `runtime.HeapAlloc` after two GCs, before vs. after building. The word
  strings (10.3 MB at 1M) are shared with the input and not counted for any implementation.
- **Workload:** words picked weighted by frequency (popular words get typed more),
  cut to a random length of 1–5 runes. Same queries for every implementation.
- **Latency:** 10,000 warm-up queries, then each query timed on its own with
  `time.Now()`. Includes allocating the result slice. GC pauses show up in max
  (the 18 ms for the query-time trie); that's real behavior, not noise to remove.
  On this machine the timer ticks every ~42 ns (every value in the precomputed rows is
  a multiple of it), so read those as "~100 ns" rather than exact.

## Decisions

### Trie vs. sorted array vs. compressed trie

**Sorted array** (words sorted by code point, binary search for both ends of the
prefix range, then a size-k heap over the range). It's the smallest by far: 24 bytes
per word, just the `Entry` itself. Its weakness is short prefixes: `b` matches tens of
thousands of words and all of them are scanned. Even so, its worst p99 (385 µs for one
letter) is hundreds of times below what a person can notice between keystrokes.

**Plain trie** (one node per rune). Finding the prefix node is fast, but there are
several nodes per word and each one is a separate heap object with pointers.
Collecting the top k still means visiting every node under the prefix, by
following pointers spread around memory. That makes the query-time trie about
**9× slower than the sorted array on 1-letter prefixes** while using 8× the memory.
The sorted array scans the same words from one contiguous block, which the CPU cache
and prefetcher handle very well.

**Radix tree** (chains of single-child nodes merged into one edge with a string label).
There are at most ~2 nodes per word no matter how long words are, so it uses **35% less
memory than the trie** and its DFS is ~2× faster (fewer nodes to visit). Its bytes per
word stay flat at ~131 from 100k to 2M words, while the trie's grow from 192 to 231
(rare words are longer and share less, so the trie adds more nodes per word). Inserting
is harder: the word can end or diverge in the middle of an edge, which splits it.

The plain trie loses to the radix tree on every measurement, so the real choice is
between the sorted array (memory) and the precomputed radix tree (speed).

### Top-k: precomputed per node vs. computed at query time

Computing at query time costs nothing extra to store, but the cost depends on how many
words share the prefix. Short prefixes are slow, which is exactly when people use
autocomplete most.

Storing the top k on every non-leaf node makes every lookup "walk the prefix, copy ≤ k
entries": **~100 ns, and the same whether the prefix is 1 or 5 letters**. On the radix
tree it costs **+41% memory** (125 → 177 MB) and **+45% build time**. Leaves don't
store a list, since it would only contain their own word.

Downsides:
- k is fixed when building. Asking for more than the built k falls back to the
  query-time search (tested in `TestK`).
- Changing a word's frequency means updating the lists on every node along its path,
  which matters if the data changes while it's being served. This project builds once
  and never updates.

### Turkish characters

- **Lowercasing must be Turkish-aware.** Turkish has dotted and dotless i as separate
  letters: `I ↔ ı` and `İ ↔ i`. `normalize.Word` uses
  `cases.Lower(language.Turkish)` from `golang.org/x/text`, so `IŞIK → ışık` and
  `İSTANBUL → istanbul`. The same function runs on the data and on what the user types,
  so they always match.
- **The data needed it.** The source list was lowercased with a function that doesn't
  know Turkish, so words starting with `İ` were never lowercased: `İşte` 195,150 sits
  next to `işte` 142,029, and `İyi` 343,130 next to `iyi` 767,278. Without merging them,
  `işte` would be ranked with less than half its real count. 2,035,629 lines become
  2,004,272 words after merging.
- **`ı` and `i` stay different letters.** `ıstanbul` (3 occurrences) does not merge
  into `istanbul`, and the prefix `ı` does not match words starting with `i`.
- **NFC normalization** makes `ğ` one code point even if it arrived as `g` + combining
  breve, so it can't miss on lookup.
- **Plain code-point order is enough** for the sorted array. Prefix search only needs
  every word with the same prefix to sit next to each other, which any character-by-
  character order guarantees. Turkish dictionary order (`ç` after `c`) only matters when
  showing a sorted list to a person, and this project never does.
- **Runes, not bytes.** The trie steps one rune at a time, and the radix tree only
  splits labels on rune boundaries. Its children are keyed by first rune, not
  first byte, because different letters can start with the same byte: `ı` and `ğ`
  are both `0xC4…`, and `ç`, `ö`, `ü` are all `0xC3…`. The brute-force test uses
  exactly those letters. Removing the rune-boundary check makes `insert` loop
  forever on them.
- **No accent folding** (typing `sehir` doesn't find `şehir`). In Turkish these are
  different letters and often different words: `kas` (muscle) / `kaş` (eyebrow),
  `cam` (glass) / `çam` (pine), `sık` (frequent) / `şık` (elegant). Folding would mix
  them up. A better design for people without a Turkish keyboard is a second, folded
  index that's only consulted when the exact prefix returns fewer than k results.
  Not built here.

## Trie experiments

Same bench, 1M words. Each row adds one change to the row above.

| Trie version | query-time memory | precomputed memory | query-time p50 / p99 (1-char) |
|---|---|---|---|
| `map[rune]*node`, map created for every node, top-k slices keep extra capacity | 415.8 MB | 562.9 MB | 4.4 ms / 10.4 ms |
| + map only created on first child, top-k lists cloned to exact size | 384.9 MB | 484.4 MB | 4.4 ms / 10.4 ms |
| + no top-k list stored on leaves | — | 468.9 MB | — |
| + children as a sorted slice of `{rune, *node}` instead of a map | **192.7 MB** | **276.8 MB** | **0.92 ms / 2.18 ms** |

- **Maps were most of the cost.** Most trie nodes have one to three children, and a Go
  map has a header plus bucket storage even for a single key. A sorted slice of
  `{rune, *node}` costs 16 bytes per child. Binary search over a few entries is also
  faster than hashing, and iterating a slice is much faster than iterating a map,
  hence the 5× faster DFS.
- **Leftover capacity matters at scale.** `allEntries[:k]` keeps the whole backing array
  (every child's list combined) alive; `slices.Clone` keeps exactly k. (`slices.Clip`
  wouldn't help: it shrinks the capacity but the big array stays.)
- **Skipping leaf lists** was a small win (3%), because the slice header stays in the
  node; only the separate allocation goes away.

## Findings

1. **Memory layout beats algorithmic shape.** The sorted array and the query-time trie
   look at the same set of matching words, but the array reads them from one contiguous
   block, and it wins by 9× on the hardest queries.
2. **Precomputation turns a variable cost into a constant one.** Precomputed lookups
   take ~100 ns at every prefix length; query-time lookups range from 5 µs to 1 ms+
   depending on how popular the prefix is.
3. **Per-node overhead dominates node-based structures.** Swapping one field type (map →
   slice) halved the trie's memory; collapsing nodes (radix) cut another third.
4. **Normalization is a correctness issue, not polish.** In the real data, 31% of
   `iyi`'s count and 58% of `işte`'s were hidden under an uppercase `İ`.
5. **For this size, the simple answer is good enough.** 22.9 MB and p99 < 0.4 ms for a
   sorted array is plenty for a human typing; the faster structures only pay off at high
   query volume.

## Possible next steps

- Store nodes in one big slice and use `int32` indices instead of pointers, to cut
  pointer overhead and GC scan time.
- A DAWG / FST (as used by Lucene) also shares word *suffixes*, not just prefixes.
- The folded (`sehir → şehir`) fallback index described above.
- Updating frequencies without rebuilding the precomputed lists.
