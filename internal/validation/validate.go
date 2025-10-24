// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package validation

import (
    "fmt"
    "os"
    errors "pqckeygen/internal/errors"
)

const MinDERKeySize = 256

// ValidateDERKey ensures the DER key file at filePath exists and meets the minimum size constraint
//
// Parameters:
//   filePath string: Path to DER file to validate
//
// Returns:
//   error:   KeyValidationError (typed) if file does not exist or is too small
func ValidateDERKey(filePath string) error {
    stat, err := os.Stat(filePath)

    if os.IsNotExist(err) {
        return &errors.KeyValidationError{errors.KeyGenError{
            Message: fmt.Sprintf("DER key file %q does not exist", filePath),
            Context: map[string]interface{}{"file": filePath},
        }}
    }

    if err != nil {
        return &errors.KeyValidationError{errors.KeyGenError{
            Message: "Could not stat DER key file",
            Context: map[string]interface{}{"file": filePath, "err": err.Error()},
        }}
    }

    if stat.Size() < MinDERKeySize {
        return &errors.KeyValidationError{errors.KeyGenError{
            Message: fmt.Sprintf("DER key file is too small (%d bytes, min %d)", stat.Size(), MinDERKeySize),
            Context: map[string]interface{}{"file": filePath, "size": stat.Size()},
        }}
    }
    
    return nil
}
