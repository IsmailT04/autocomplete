// Command repl is an interactive prompt for trying suggestions by hand.
package main

import (
	"flag"
)

func main() {
	_ = flag.String("data", "tr_full.txt", "FrequencyWords-style word list")
	_ = flag.String("impl", "trie", "sortedarray | trie | radix")
	flag.Parse()

	// TODO: load data, build the chosen implementation, then loop:
	// read a line (bufio.Scanner on os.Stdin), normalize it, print Suggest(prefix, 10)
	// and how long the lookup took.
}
