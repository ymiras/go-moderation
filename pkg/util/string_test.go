package util

import (
	"testing"
)

func TestMask(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		start    int
		end      int
		expected string
	}{
		{"basic mask", "hello world", 0, 5, "***** world"},
		{"end beyond length", "hi", 0, 100, "**"},
		{"negative start", "hello", -5, 5, "*****"},
		{"start equals end", "hello", 3, 3, "hello"},
		{"middle mask", "hello world", 6, 11, "hello *****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mask(tt.text, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("Mask(%q, %d, %d) = %q, want %q", tt.text, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		words    []string
		expected bool
	}{
		{"match found", "hello world", []string{"hello", "goodbye"}, true},
		{"no match", "hello world", []string{"goodbye", "ciao"}, false},
		{"empty words", "hello", []string{}, false},
		{"empty text", "", []string{"hello"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsAny(tt.text, tt.words)
			if result != tt.expected {
				t.Errorf("ContainsAny(%q, %v) = %v, want %v", tt.text, tt.words, result, tt.expected)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"full-width to half-width", " hｅllo ", " hello "},
		{"ideographic space", "hello　world", "hello world"},
		{"zero-width space", "hello\u200bworld", "helloworld"},
		{"already normalized", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFirstMatchIndex(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		words         []string
		expectedIndex int
		expectedWord  string
	}{
		{"find first", "hello world", []string{"world", "hello"}, 0, "hello"},
		{"only second matches", "goodbye world", []string{"foo", "world"}, 8, "world"},
		{"no match", "hello world", []string{"foo", "bar"}, -1, ""},
		{"empty words", "hello", []string{}, -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, word := FirstMatchIndex(tt.text, tt.words)
			if idx != tt.expectedIndex {
				t.Errorf("FirstMatchIndex(%q, %v) index = %d, want %d", tt.text, tt.words, idx, tt.expectedIndex)
			}
			if word != tt.expectedWord {
				t.Errorf("FirstMatchIndex(%q, %v) word = %q, want %q", tt.text, tt.words, word, tt.expectedWord)
			}
		})
	}
}

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"ascii only", "hello", true},
		{"with spaces", "hello world", true},
		{"with unicode", "hėllo", false},
		{"with emoji", "hello👋", false},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsASCII(tt.text)
			if result != tt.expected {
				t.Errorf("IsASCII(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}
