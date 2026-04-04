package model

import (
	"testing"
)

func TestActionType_Values(t *testing.T) {
	tests := []struct {
		name     string
		value    ActionType
		expected string
	}{
		{"ActionPass", ActionPass, "pass"},
		{"ActionBlock", ActionBlock, "block"},
		{"ActionMask", ActionMask, "mask"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("ActionType = %v, want %v", tt.value, tt.expected)
			}
		})
	}
}

func TestPointType_Values(t *testing.T) {
	tests := []struct {
		name     string
		value    PointType
		expected string
	}{
		{"PointInput", PointInput, "input"},
		{"PointOutput", PointOutput, "output"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("PointType = %v, want %v", tt.value, tt.expected)
			}
		})
	}
}

func TestSeverity_Values(t *testing.T) {
	tests := []struct {
		name     string
		value    Severity
		expected string
	}{
		{"SeverityLow", SeverityLow, "low"},
		{"SeverityMedium", SeverityMedium, "medium"},
		{"SeverityHigh", SeverityHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("Severity = %v, want %v", tt.value, tt.expected)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	err := &Error{
		Code:    "E001",
		Message: "test error",
	}

	expected := "[E001] test error"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %v, want %v", got, expected)
	}
}

func TestNewError(t *testing.T) {
	err := NewError("E002", "word bank error")

	if err.Code != "E002" {
		t.Errorf("Code = %v, want %v", err.Code, "E002")
	}
	if err.Message != "word bank error" {
		t.Errorf("Message = %v, want %v", err.Message, "word bank error")
	}
}

func TestErrConfigLoad(t *testing.T) {
	err := ErrConfigLoad("missing file")

	if err.Code != ErrCodeConfigLoad {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeConfigLoad)
	}
	if err.Message != "missing file" {
		t.Errorf("Message = %v, want %v", err.Message, "missing file")
	}
}

func TestErrWordBankLoad(t *testing.T) {
	err := ErrWordBankLoad("invalid format")

	if err.Code != ErrCodeWordBankLoad {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeWordBankLoad)
	}
	if err.Message != "invalid format" {
		t.Errorf("Message = %v, want %v", err.Message, "invalid format")
	}
}

func TestErrInvalidInput(t *testing.T) {
	err := ErrInvalidInput("text too long")

	if err.Code != ErrCodeInvalidInput {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeInvalidInput)
	}
	if err.Message != "text too long" {
		t.Errorf("Message = %v, want %v", err.Message, "text too long")
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify error codes are unique
	codes := []string{ErrCodeConfigLoad, ErrCodeWordBankLoad, ErrCodeInvalidInput}
	seen := make(map[string]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate error code: %v", code)
		}
		seen[code] = true
	}
}
