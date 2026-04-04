package model

import "fmt"

// Error codes as string constants with prefix E
const (
	ErrCodeConfigLoad   = "E001" // Configuration loading error
	ErrCodeWordBankLoad = "E002" // Word bank loading error
	ErrCodeInvalidInput = "E003" // Invalid input error
)

// Error represents a domain error with code and message
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError creates a new domain error
func NewError(code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// ErrConfigLoad creates a configuration loading error
func ErrConfigLoad(message string) *Error {
	return NewError(ErrCodeConfigLoad, message)
}

// ErrWordBankLoad creates a word bank loading error
func ErrWordBankLoad(message string) *Error {
	return NewError(ErrCodeWordBankLoad, message)
}

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(message string) *Error {
	return NewError(ErrCodeInvalidInput, message)
}
