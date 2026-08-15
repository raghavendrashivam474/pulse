package errors

import "fmt"

// Kind classifies the type of error Aryntra Aayam encountered.
type Kind int

const (
	// KindUser represents an invalid input or argument from the user.
	KindUser Kind = iota

	// KindEnvironment represents a system or filesystem failure.
	KindEnvironment

	// KindInternal represents an unexpected application-level failure.
	KindInternal
)

// AayamError is the application-level error type.
type AayamError struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *AayamError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AayamError) Unwrap() error {
	return e.Cause
}

// User constructs a user-input error.
func User(message string) *AayamError {
	return &AayamError{Kind: KindUser, Message: message}
}

// UserWrap constructs a user-input error with an underlying cause.
func UserWrap(message string, cause error) *AayamError {
	return &AayamError{Kind: KindUser, Message: message, Cause: cause}
}

// Environment constructs a system or filesystem error.
func Environment(message string, cause error) *AayamError {
	return &AayamError{Kind: KindEnvironment, Message: message, Cause: cause}
}

// Internal constructs an unexpected application-level error.
func Internal(message string, cause error) *AayamError {
	return &AayamError{Kind: KindInternal, Message: message, Cause: cause}
}
