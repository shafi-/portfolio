package integration

import (
	"errors"
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func NewError(code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

var (
	ErrNotFound         = errors.New("integration not found")
	ErrAlreadyInstalled = errors.New("integration already installed")
	ErrNotInstalled     = errors.New("integration not installed")
	ErrIncompatible     = errors.New("integration incompatible with engine version")
	ErrUpgradeFailed    = errors.New("integration upgrade failed")
	ErrRollbackFailed   = errors.New("integration rollback failed")
	ErrStoreUnavailable = errors.New("knowledge store unavailable")
)

const (
	ErrCodeNotFound         = "INTEGRATION_NOT_FOUND"
	ErrCodeAlreadyInstalled = "INTEGRATION_ALREADY_INSTALLED"
	ErrCodeNotInstalled     = "INTEGRATION_NOT_INSTALLED"
	ErrCodeIncompatible     = "INTEGRATION_INCOMPATIBLE"
	ErrCodeUpgradeFailed    = "INTEGRATION_UPGRADE_FAILED"
	ErrCodeRollbackFailed   = "INTEGRATION_ROLLBACK_FAILED"
	ErrCodeStoreUnavailable = "STORE_UNAVAILABLE"
	ErrCodePermissionDenied = "PERMISSION_DENIED"
)
