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

Machine: TODO. Go 1.27. n = 1,000,000 words, k = 10.

| Implementation | Build time | Memory | p50 | p99 | p99 (1-char prefix) |
|---|---|---|---|---|---|
| sortedarray | | | | | |
| trie / query-time | | | | | |
| trie / precomputed | | | | | |
| radix / query-time | | | | | |
| radix / precomputed | | | | | |

## What I learned
TODO
