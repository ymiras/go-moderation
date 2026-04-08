package storage

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ymiras/go-moderation/internal/model"
)

// WordBank defines the interface for keyword storage.
type WordBank interface {
	Load(path string) error
	Contains(word string) (bool, *model.Keyword)
	Words() []*model.Keyword
	Size() int
}

// inMemoryWordBank implements WordBank with in-memory storage.
type inMemoryWordBank struct {
	mu    sync.RWMutex
	words map[string]*model.Keyword
	index map[string][]*model.Keyword // type -> keywords
}

// NewWordBank creates a new in-memory word bank.
func NewWordBank() WordBank {
	return &inMemoryWordBank{
		words: make(map[string]*model.Keyword),
		index: make(map[string][]*model.Keyword),
	}
}

// Load reads keywords from a CSV file.
func (wb *inMemoryWordBank) Load(path string) error {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open word bank file: %w", err)
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 4
	reader.Comment = '#'

	// Clear existing data
	wb.words = make(map[string]*model.Keyword)
	wb.index = make(map[string][]*model.Keyword)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		// Skip empty rows
		if len(record) == 0 || strings.TrimSpace(record[0]) == "" {
			continue
		}

		keyword := &model.Keyword{
			Word: strings.TrimSpace(record[0]),
		}

		// Optional columns with defaults
		if len(record) > 1 {
			keyword.Type = strings.TrimSpace(record[1])
		}
		if len(record) > 2 {
			keyword.Severity = model.Severity(strings.TrimSpace(record[2]))
		}
		if len(record) > 3 {
			keyword.Action = model.ActionType(strings.TrimSpace(record[3]))
		}

		wb.words[keyword.Word] = keyword
		if keyword.Type != "" {
			wb.index[keyword.Type] = append(wb.index[keyword.Type], keyword)
		}
	}

	return nil
}

// Contains checks if a word exists in the word bank.
func (wb *inMemoryWordBank) Contains(word string) (bool, *model.Keyword) {
	wb.mu.RLock()
	defer wb.mu.RUnlock()

	kw, ok := wb.words[word]
	return ok, kw
}

// Words returns all keywords.
func (wb *inMemoryWordBank) Words() []*model.Keyword {
	wb.mu.RLock()
	defer wb.mu.RUnlock()

	result := make([]*model.Keyword, 0, len(wb.words))
	for _, kw := range wb.words {
		result = append(result, kw)
	}
	return result
}

// Size returns the number of keywords.
func (wb *inMemoryWordBank) Size() int {
	wb.mu.RLock()
	defer wb.mu.RUnlock()
	return len(wb.words)
}
