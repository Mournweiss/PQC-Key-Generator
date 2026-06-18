// SPDX-FileCopyrightText: 2026 Maxim Selin <info@mournweiss.ru>
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

// DERWorker implements KeyFormatWorker for generation and validation of ASN.1 DER keys.
//
// Used by the format registry; responsible for producing and validating DER-formatted keys only.
type DERWorker struct {
    templates.BaseFormatWorker
}

// init registers DERWorker with the format registry for the "DER" key type.
func init() {
    RegisterFormatWorker("DER", &DERWorker{})
}

// Generate creates and exports a key in DER format using OpenSSL.
// If KEYGEN_KEYPAIR=true in env, preserves the PEM file for use as public key (does not remove PEM after export). Otherwise, deletes PEM.
//
// Params:
//   algorithm   (string): OpenSSL algorithm name for key generation.
//   baseOutPath (string): Base path (without extension) for resulting files.
//
// Returns:
//   (string): Absolute path to the created DER file (.der extension).
//   (error):  If generation or export fails, or DER file is invalid/missing.
func (w *DERWorker) Generate(algorithm string, baseOutPath string) (string, error) {
    pemFile := baseOutPath + ".pem"
    derFile := baseOutPath + ".der"
    preservePEM := false
    if val, ok := os.LookupEnv("KEYGEN_KEYPAIR"); ok && val == "true" {
        preservePEM = true
    }
    if !utils.FileExists(pemFile) {
        genCmd := exec.Command("openssl", "genpkey", "-provider", "default", "-algorithm", algorithm, "-out", pemFile)
        genOut, err := genCmd.CombinedOutput()
        if err != nil || !utils.FileExists(pemFile) {
            return "", &errors.PQCNotSupportedError{errors.KeyGenError{
                Message: fmt.Sprintf("Key generation via openssl failed for %s (DER)", algorithm),
                Context: map[string]interface{}{ "cmd": genCmd.String(), "output": string(genOut), "path": pemFile, "algorithm": algorithm },
            }}
        }
    }
    derCmd := exec.Command("openssl", "pkey", "-in", pemFile, "-outform", "DER", "-out", derFile)
    derOut, err := derCmd.CombinedOutput()
    if err != nil || !utils.FileExists(derFile) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("DER export failed for %s", algorithm),
            Context: map[string]interface{}{ "cmd": derCmd.String(), "output": string(derOut), "path": derFile, "algorithm": algorithm },
        }}
    }
    if !preservePEM {
        _ = os.Remove(pemFile) // Only remove PEM if we're NOT in keypair mode
    }
    return derFile, nil
}

// Validate checks the specified DER file for existence and reasonable minimum size to confirm validity.
//
// Params:
//   outPath (string): Path to the DER file to validate.
//
// Returns:
//   (error):  If missing, too small, not a valid DER; nil otherwise.
func (w *DERWorker) Validate(outPath string) error {
    stat, err := os.Stat(outPath)
    if os.IsNotExist(err) {
        return fmt.Errorf("DER key file %q does not exist", outPath)
    }
    if err != nil {
        return fmt.Errorf("Could not stat DER key file: %v", err)
    }
    if stat.Size() < 256 {
        return fmt.Errorf("DER key file is too small (%d bytes, min %d)", stat.Size(), 256)
    }
    return nil
}
