package dataset

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"autocomplete/internal/suggest"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "words.txt")
	data := "" +
		"istanbul 100\n" +
		"İstanbul 20\n" + // normalizes to "istanbul": counts are summed
		"ışık 50\n" +
		"IŞIK 5\n" + // Turkish lowercasing: I → ı, so this merges with ışık
		"\n" + // empty line
		"bozuk\n" + // no count
		"sayı abc\n" + // count is not a number
		"az 50\n" +
		"ay 50\n" +
		"nadir 1\n" // below minFreq
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := []suggest.Entry{
		{Word: "istanbul", Freq: 120},
		{Word: "ışık", Freq: 55},
		{Word: "ay", Freq: 50}, // equal Freq: word order
		{Word: "az", Freq: 50},
	}
	if !slices.Equal(got, want) {
		t.Errorf("Load = %v, want %v", got, want)
	}
}

func TestTopN(t *testing.T) {
	es := []suggest.Entry{{Word: "a", Freq: 3}, {Word: "b", Freq: 2}, {Word: "c", Freq: 1}}
	if got := TopN(es, 2); !slices.Equal(got, es[:2]) {
		t.Errorf("TopN(2) = %v", got)
	}
	if got := TopN(es, 10); !slices.Equal(got, es) {
		t.Errorf("TopN(10) = %v", got)
	}
	if got := TopN(es, 0); got != nil {
		t.Errorf("TopN(0) = %v, want nil", got)
	}
	// TopN must copy: changing the result must not change the input.
	TopN(es, 1)[0].Word = "x"
	if es[0].Word != "a" {
		t.Error("TopN modified its input")
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.txt"), 1); err == nil {
		t.Error("expected an error for a missing file")
	}
}
