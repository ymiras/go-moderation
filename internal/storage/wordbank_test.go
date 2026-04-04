package storage

import (
	"os"
	"testing"

	"github.com/ymiras/dify-moderation/internal/model"
)

func TestNewWordBank(t *testing.T) {
	wb := NewWordBank()
	if wb == nil {
		t.Fatal("NewWordBank() returned nil")
	}
}

func TestWordBankLoad(t *testing.T) {
	// Create a temporary CSV file
	content := `hello,political,high,block
world,ad,medium,mask
test,political,low,pass`

	tmpfile, err := os.CreateTemp("", "wordbank-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	wb := NewWordBank()
	err = wb.Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if wb.Size() != 3 {
		t.Errorf("Size() = %d, want 3", wb.Size())
	}
}

func TestWordBankContains(t *testing.T) {
	content := `hello,political,high,block
world,ad,medium,mask`

	tmpfile, err := os.CreateTemp("", "wordbank-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	wb := NewWordBank()
	wb.Load(tmpfile.Name())

	found, kw := wb.Contains("hello")
	if !found {
		t.Error("Contains(hello) = false, want true")
	}
	if kw == nil || kw.Word != "hello" {
		t.Errorf("Contains(hello) keyword = %v, want hello", kw)
	}
	if kw.Severity != model.SeverityHigh {
		t.Errorf("kw.Severity = %v, want high", kw.Severity)
	}

	found, _ = wb.Contains("notexist")
	if found {
		t.Error("Contains(notexist) = true, want false")
	}
}

func TestWordBankWords(t *testing.T) {
	content := `hello,political,high,block
world,ad,medium,mask`

	tmpfile, err := os.CreateTemp("", "wordbank-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	wb := NewWordBank()
	wb.Load(tmpfile.Name())

	words := wb.Words()
	if len(words) != 2 {
		t.Errorf("len(Words()) = %d, want 2", len(words))
	}
}

func TestWordBankLoadFileNotFound(t *testing.T) {
	wb := NewWordBank()
	err := wb.Load("nonexistent.csv")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}
