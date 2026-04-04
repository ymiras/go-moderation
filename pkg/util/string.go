package util

import (
	"strings"
	"unicode"
)

// Mask replaces characters in text from start to end with asterisks.
func Mask(text string, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}
	if start >= end {
		return text
	}

	prefix := text[:start]
	suffix := text[end:]
	masks := strings.Repeat("*", end-start)
	return prefix + masks + suffix
}

// ContainsAny returns true if text contains any of the words.
func ContainsAny(text string, words []string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// Normalize converts full-width characters to half-width for comparison.
func Normalize(text string) string {
	var result strings.Builder
	for _, r := range text {
		// Convert full-width ASCII to half-width
		if r >= 0xFF01 && r <= 0xFF5E {
			result.WriteRune(r - 0xFEE0)
		} else if r == 0x3000 { // ideographic space
			result.WriteRune(' ')
		} else if r == 0x200B { // zero-width space
			// skip
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FirstMatchIndex returns the index and word of the first match found in text.
// Returns -1 and empty string if no match is found.
func FirstMatchIndex(text string, words []string) (int, string) {
	lowestIndex := -1
	matchedWord := ""

	for _, word := range words {
		idx := strings.Index(text, word)
		if idx != -1 {
			if lowestIndex == -1 || idx < lowestIndex {
				lowestIndex = idx
				matchedWord = word
			}
		}
	}

	return lowestIndex, matchedWord
}

// IsASCII returns true if all characters in text are ASCII.
func IsASCII(text string) bool {
	for _, r := range text {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}
