// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT
//
package pqc

import (
    "fmt"
    cfgmod "pqckeygen/internal/config"
    "pqckeygen/internal/errors"
    fmtmod "pqckeygen/internal/formats"
)

// GenerateKey performs cryptographic key generation using the registered format worker (supports keypair mode).
//
// Params:
//   cfg        (*config.Config): Loaded configuration with all parameters (algorithm, mode, format...)
//   outBasePath (string): Output base filesystem path for storing generated key(s).
//
// Returns:
//   (string):  If keypair mode is enabled, returns comma-separated absolute PEM and DER paths; else returns absolute path to generated key.
//   (error):   If format/worker is not registered or generation/validation fails; nil if success.
func GenerateKey(cfg *cfgmod.Config, outBasePath string) (string, error) {
    if cfg.KeyPairMode {
        // Keypair mode: PEM + DER output with original PEM retained
        // Generate PEM private key using PEMWorker directly
        pemWorker, ok := fmtmod.GetFormatWorker("PEM")
        if !ok || pemWorker == nil {
            return "", &errors.PQCNotSupportedError{errors.KeyGenError{
                Message: "No PEM worker registered in the format registry.",
                Context: map[string]interface{}{"mode": "keypair"},
            }}
        }
        generatedPem, err := pemWorker.Generate(cfg.Algorithm, outBasePath)
        if err != nil {
            return "", err
        }
        // Validate PEM
        if vErr := pemWorker.Validate(generatedPem); vErr != nil {
            return "", vErr
        }
        // Export DER from PEM using DERWorker
        derWorker, ok := fmtmod.GetFormatWorker("DER")
        if !ok || derWorker == nil {
            return "", &errors.PQCNotSupportedError{errors.KeyGenError{
                Message: "No DER worker registered in the format registry.",
                Context: map[string]interface{}{"mode": "keypair"},
            }}
        }
        generatedDer, derr := derWorker.Generate(cfg.Algorithm, outBasePath)
        if derr != nil {
            return "", derr
        }
        // Validate DER
        if vErr := derWorker.Validate(generatedDer); vErr != nil {
            return "", vErr
        }
        // Output both absolute paths, comma-separated
        return fmt.Sprintf("%s,%s", generatedPem, generatedDer), nil
    } else {
        // Standard mode — direct call via registry/worker as before
        if !cfgmod.IsSupportedFormat(string(cfg.Format)) {
            return "", &errors.PQCNotSupportedError{errors.KeyGenError{
                Message: fmt.Sprintf("Unsupported format: %s", string(cfg.Format)),
                Context: map[string]interface{}{"algorithm": cfg.Algorithm, "format": cfg.Format},
            }}
        }
        worker, ok := fmtmod.GetFormatWorker(string(cfg.Format))
        if !ok || worker == nil {
            return "", &errors.PQCNotSupportedError{errors.KeyGenError{
                Message: fmt.Sprintf("No worker registered for format: %s (registry not found)", string(cfg.Format)),
                Context: map[string]interface{}{"algorithm": cfg.Algorithm, "format": cfg.Format},
            }}
        }
        outPathRes, err := worker.Generate(cfg.Algorithm, outBasePath)
        if err != nil {
            return "", err
        }
        if vErr := worker.Validate(outPathRes); vErr != nil {
            return "", vErr
        }
        return outPathRes, nil
    }
}
