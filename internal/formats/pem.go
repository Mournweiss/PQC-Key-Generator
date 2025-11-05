// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package formats

import (
    "fmt"
    "os"
    "os/exec"
    "pqckeygen/internal/errors"
    "pqckeygen/internal/utils"
    "pqckeygen/internal/formats/templates"
)

// PEMWorker implements KeyFormatWorker for generation and validation of PEM encoded private keys.
//
// Registered to the dynamic format registry, provides all PEM operations for keygen.
type PEMWorker struct {
    templates.BaseFormatWorker
}

// init registers PEMWorker with the format registry for the "PEM" key type.
func init() {
    RegisterFormatWorker("PEM", &PEMWorker{})
}

// Generate creates and exports a key in PEM format using OpenSSL.
//
// Params:
//   algorithm (string): OpenSSL algorithm name.
//   baseOutPath (string): Base path (no extension) for resulting PEM file.
//
// Returns:
//   (string): Final PEM file path (always with .pem extension).
//   (error):  On any OpenSSL/IO error or unsuccessful key export.
func (w *PEMWorker) Generate(algorithm string, baseOutPath string) (string, error) {
    pemFile := baseOutPath + ".pem"
    genCmd := exec.Command("openssl", "genpkey", "-provider", "default", "-algorithm", algorithm, "-out", pemFile)
    genOut, err := genCmd.CombinedOutput()
    if err != nil || !utils.FileExists(pemFile) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("Key generation via openssl failed for %s (PEM)", algorithm),
            Context: map[string]interface{}{ "cmd": genCmd.String(), "output": string(genOut), "path": pemFile, "algorithm": algorithm },
        }}
    }
    return pemFile, nil
}

// Validate performs PEM-specific validation: checks file existence, size, and header.
//
// Params:
//   outPath (string): Path to the PEM file to validate.
//
// Returns:
//   (error):  If file does not exist, empty, or missing PEM header; nil if valid.
func (w *PEMWorker) Validate(outPath string) error {
    stat, err := os.Stat(outPath)
    if err != nil {
        return fmt.Errorf("PEM key file does not exist: %v", err)
    }
    if stat.Size() == 0 {
        return fmt.Errorf("PEM key file is empty")
    }
    f, err := os.Open(outPath)
    if err != nil {
        return fmt.Errorf("Could not open PEM key file: %v", err)
    }
    defer f.Close()
    header := make([]byte, 32)
    n, err := f.Read(header)
    if err != nil {
        return fmt.Errorf("Could not read PEM key file: %v", err)
    }
    if n == 0 || string(header[:27]) != "-----BEGIN PRIVATE KEY-----" {
        return fmt.Errorf("PEM header missing or invalid")
    }
    return nil
}
