package errors_test

import (
	"fmt"
	"testing"

	pulseErrors "pulse/internal/errors"
)

func TestUserError(t *testing.T) {
	err := pulseErrors.User("invalid path supplied")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Kind != pulseErrors.KindUser {
		t.Errorf("expected KindUser, got %v", err.Kind)
	}
	if err.Error() != "invalid path supplied" {
		t.Errorf("unexpected message: %q", err.Error())
	}
}

func TestUserWrapError(t *testing.T) {
	cause := fmt.Errorf("no such file or directory")
	err := pulseErrors.UserWrap("target path does not exist: ./missing", cause)
	if err.Kind != pulseErrors.KindUser {
		t.Errorf("expected KindUser, got %v", err.Kind)
	}
	if err.Unwrap() != cause {
		t.Error("expected Unwrap to return the cause")
	}
	expected := "target path does not exist: ./missing: no such file or directory"
	if err.Error() != expected {
		t.Errorf("unexpected error string: %q", err.Error())
	}
}

func TestEnvironmentError(t *testing.T) {
	cause := fmt.Errorf("permission denied")
	err := pulseErrors.Environment("cannot read project directory", cause)
	if err.Kind != pulseErrors.KindEnvironment {
		t.Errorf("expected KindEnvironment, got %v", err.Kind)
	}
	if err.Unwrap() != cause {
		t.Error("expected Unwrap to return the cause")
	}
}

func TestInternalError(t *testing.T) {
	cause := fmt.Errorf("unexpected nil state")
	err := pulseErrors.Internal("snapshot construction failed", cause)
	if err.Kind != pulseErrors.KindInternal {
		t.Errorf("expected KindInternal, got %v", err.Kind)
	}
	if err.Unwrap() != cause {
		t.Error("expected Unwrap to return the cause")
	}
}

func TestErrorWithoutCause(t *testing.T) {
	err := pulseErrors.User("too many arguments")
	if err.Unwrap() != nil {
		t.Error("expected nil cause for error constructed without cause")
	}
}
