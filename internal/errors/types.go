// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package errors

import (
    "fmt"
)

// KeyGenError is the base error type for all key generation-related failures
//
// Fields:
//   Message string: Human-readable summary of the failure
//   Context map[string]interface{}: Additional context for logging/troubleshooting
type KeyGenError struct {
    Message string
    Context map[string]interface{}
}

// Error returns the string representation of the KeyGenError
func (e *KeyGenError) Error() string {
    return e.Message
}

// LogContext returns a detailed log line with the error and its context
func (e *KeyGenError) LogContext() string {
    return fmt.Sprintf("KeyGenError %s | Context: %v", e.Message, e.Context)
}

// PQCNotSupportedError signals a lack of support for the requested PQC algorithm or operation
type PQCNotSupportedError struct {
    KeyGenError
}

// KeyValidationError signals failure to validate a DER key file
type KeyValidationError struct {
    KeyGenError
}
