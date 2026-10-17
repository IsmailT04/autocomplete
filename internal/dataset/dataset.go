// Package dataset loads a FrequencyWords-style list ("word count" per line).
package dataset

import (
	"autocomplete/internal/normalize"
	"autocomplete/internal/suggest"
	"bufio"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Load reads path, normalizes every word, merges duplicates that collapse to the
// same normalized form (summing their counts), and drops words with
// Freq < minFreq. Malformed lines are skipped.
//
// The result is sorted by Freq descending, ties by Word ascending (the same
// order as the suggest.Suggester contract). TopN and the precomputed top-k
// trie rely on this order.
func Load(path string, minFreq uint64) ([]suggest.Entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	counts := make(map[string]uint64)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		//skip empty lines or comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue //Malformed line
		}

		word := normalize.Word(fields[0])
		if word == "" {
			continue
		}

		freq, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue //malformed freq
		}

		counts[word] += freq
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Filter by minFreq and collect into a slice
	entries := make([]suggest.Entry, 0, len(counts))
	for word, freq := range counts {
		if freq >= minFreq {
			entries = append(entries, suggest.Entry{
				Word: word,
				Freq: freq,
			})
		}
	}

	// Sort descending by frequency, with alphabetical tie-breaking for determinism
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Freq == entries[j].Freq {
			return entries[i].Word < entries[j].Word
		}
		return entries[i].Freq > entries[j].Freq
	})

	return entries, nil
}

// TopN returns a copy of the n most frequent entries. Useful for benchmarking
// at exact sizes (100k, 500k, 1M, ...). entries must already be in the order
// Load returns; TopN does not sort.
func TopN(entries []suggest.Entry, n int) []suggest.Entry {
	if n <= 0 || len(entries) == 0 {
		return nil
	}

	if n > len(entries) {
		n = len(entries)
	}

	// Copy only the first n, so the caller can drop the full list and its memory.
	return slices.Clone(entries[:n])
}
