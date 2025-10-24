// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package utils

import (
    "crypto/rand"
    "encoding/hex"
    "os"
)

// RandomHex generates a cryptographically secure random hex string
//
// Parameters:
//   n int: Number of bytes to generate (result will be 2*n hex chars)
//
// Returns:
//   string: Hex-encoded random string
//   error:  If random source fails
func RandomHex(n int) (string, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

// FileExists checks if the given path exists and is not a directory
//
// Parameters:
//   path string: File path to check
//
// Returns:
//   bool: true if file exists and is not a directory, false otherwise
func FileExists(path string) bool {
    info, err := os.Stat(path)
    return err == nil && !info.IsDir()
}
