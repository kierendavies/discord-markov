package bot

import "strings"

const (
	chainLenSave = 5
	chainLenGen  = 2
)
const tokenSeparator = " "
const (
	stx = "\x02"
	etx = "\x03"
)

func tokenSequences(msg string) []string {
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
