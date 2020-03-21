package bot

import (
	"math/rand"
	"strings"
)

const (
	chainLenSave = 5
	chainLenGen  = 3
)
const tokenSeparator = " "
const (
	stx = "\x02"
	etx = "\x03"
)

func tokenChains(msg string) []string {
	tokens := strings.Split(msg, tokenSeparator)
	// Prepend start-of-text
	tokens = append([]string{stx}, tokens...)
	// Append end-of-text
	tokens = append(tokens, etx)

	tokenSeqs := make([]string, 0)
	for start := 0; start < len(tokens)-1; start++ {
		for l := 2; l <= chainLenSave+1; l++ {
			end := start + l
			if end > len(tokens) {
				break
			}

			ts := strings.Join(tokens[start:end], tokenSeparator)
			tokenSeqs = append(tokenSeqs, ts)
		}
	}

	return tokenSeqs
}

func weightedChoice(choices map[string]uint64) string {
	if len(choices) == 0 {
		panic("choices is empty")
	}

	keys := make([]string, 0, len(choices))
	cumsums := make([]uint64, 0, len(choices))
	var sum uint64 = 0

	for k, v := range choices {
		keys = append(keys, k)
		sum += v
		cumsums = append(cumsums, sum)
	}

	r := uint64(rand.Int63n(int64(sum)))
	for i, cs := range cumsums {
		if r < cs {
			return keys[i]
		}
	}

	panic("something went wrong with weightedChoice")
}
