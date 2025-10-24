package errors

import (
    "fmt"
)

type KeyGenError struct {
    Message string
    Context map[string]interface{}
}

func (e *KeyGenError) Error() string {
    return e.Message
}

func (e *KeyGenError) LogContext() string {
    return fmt.Sprintf("[KeyGenError] %s | Context: %v", e.Message, e.Context)
}

type PQCNotSupportedError struct {
    KeyGenError
}

type KeyValidationError struct {
    KeyGenError
}
