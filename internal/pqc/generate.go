// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package pqc

import (
    "fmt"
    "os/exec"
    "strings"
    cfgmod "pqckeygen/internal/config"
    "pqckeygen/internal/errors"
    fmtmod "pqckeygen/internal/formats"
)

// GenerateKey is the public factory for cryptographic key generation and validation (format-agnostic).
//
// Params:
//   algorithm (string): Name of the algorithm from OpenSSL's list.
//   format    (string): Desired key export format (must be registered).
//   outPath   (string): Output filesystem path for the generated key file.
//
// Returns:
//   (string): Path to the successfully generated key file.
//   (error):  If generation or validation fails, details as error.
func GenerateKey(algorithm string, format string, outPath string) (string, error) {
    if !cfgmod.IsSupportedFormat(format) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("Unsupported format: %s", format),
            Context: map[string]interface{}{"algorithm": algorithm, "format": format},
        }}
    }
    listCmd := exec.Command("openssl", "list", "-public-key-algorithms", "-provider", "default")
    out, err := listCmd.CombinedOutput()
    if err != nil {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: "Failed to query OpenSSL for supported algorithms",
            Context: map[string]interface{}{"output": string(out), "err": err.Error()},
        }}
    }
    if !strings.Contains(strings.ToUpper(string(out)), strings.ToUpper(algorithm)) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("%s not found in OpenSSL supported list", algorithm),
            Context: map[string]interface{}{ "output": string(out), "algorithm": algorithm},
        }}
    }
    worker, ok := fmtmod.GetFormatWorker(format)
    if !ok || worker == nil {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("No worker registered for format: %s (registry not found)", format),
            Context: map[string]interface{}{"algorithm": algorithm, "format": format},
        }}
    }
    outPathRes, err := worker.Generate(algorithm, outPath)
    if err != nil {
        return "", err
    }
    if vErr := worker.Validate(outPathRes); vErr != nil {
        return "", vErr
    }
    return outPathRes, nil
}
