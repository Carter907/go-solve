package word_count

import "strings"

// WordCount counts the occurrences of each word in a string.
func WordCount(s string) map[string]int {
	counts := make(map[string]int)
	words := strings.Fields(s)
	for _, word := range words {
		counts[word]++
	}
	return counts
}
