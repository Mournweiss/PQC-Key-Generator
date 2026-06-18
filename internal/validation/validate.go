// SPDX-FileCopyrightText: 2026 Maxim Selin <info@mournweiss.ru>
//
// SPDX-License-Identifier: MIT

package validation

import (
    "fmt"
    "pqckeygen/internal/formats"
)

// ValidateKeyByFormat performs validation using the registered format worker for the given key type.
//
// Params:
//   format   (string): Format string (must match a registered worker, e.g., "DER", "PEM").
//   filePath (string): Path to the key file to validate.
//
// Returns:
//   (error):  If format is not registered or validation fails, with error describing the reason; nil if valid.
func ValidateKeyByFormat(format string, filePath string) error {
    worker, ok := formats.GetFormatWorker(format)
    if !ok || worker == nil {
        return fmt.Errorf("validation failed: no such key format registered: %s", format)
    }
    return worker.Validate(filePath)
}
