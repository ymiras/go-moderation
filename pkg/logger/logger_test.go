package logger

import (
	"testing"
)

func TestNewProduction(t *testing.T) {
	log, err := NewProduction()
	if err != nil {
		t.Fatalf("NewProduction() error = %v", err)
	}
	if log == nil {
		t.Fatal("NewProduction() returned nil logger")
	}
	_ = log.Sync() // cleanup
}

func TestNewDevelopment(t *testing.T) {
	log, err := NewDevelopment()
	if err != nil {
		t.Fatalf("NewDevelopment() error = %v", err)
	}
	if log == nil {
		t.Fatal("NewDevelopment() returned nil logger")
	}
	_ = log.Sync() // cleanup
}

func TestNewJSON(t *testing.T) {
	log, err := NewJSON("info")
	if err != nil {
		t.Fatalf("NewJSON() error = %v", err)
	}
	if log == nil {
		t.Fatal("NewJSON() returned nil logger")
	}
	_ = log.Sync() // cleanup
}

func TestNewJSON_InvalidLevel(t *testing.T) {
	_, err := NewJSON("invalid_level")
	if err == nil {
		t.Error("NewJSON() expected error for invalid level")
	}
}
