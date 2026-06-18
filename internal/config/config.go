// SPDX-FileCopyrightText: 2026 Maxim Selin <info@mournweiss.ru>
//
// SPDX-License-Identifier: MIT

package config

import (
    "fmt"
    "os"
    "strings"
    fmtmod "pqckeygen/internal/formats"
)

// KeyFormat is a centralized type for all supported key formats.
type KeyFormat string

// GetSupportedKeyFormats dynamically returns all key formats currently registered at runtime.
//
// Returns:
//   ([]KeyFormat): List of all supported formats, as discovered from the registry
func GetSupportedKeyFormats() []KeyFormat {
    keys := fmtmod.ListRegisteredFormats()
    res := make([]KeyFormat, 0, len(keys))
    for _, k := range keys {
        res = append(res, KeyFormat(strings.ToUpper(k)))
    }
    return res
}

// IsSupportedFormat checks if the provided key format string corresponds to a registered format (case-insensitive).
//
// Params:
//   format (string): Format string, case-insensitive (e.g. "DER", "PEM")
//
// Returns:
//   (bool): true only if the format is supported by the registry at runtime.
func IsSupportedFormat(format string) bool {
    for _, kf := range GetSupportedKeyFormats() {
        if strings.EqualFold(string(kf), format) {
            return true
        }
    }
    return false
}

// Config holds all environment-driven runtime parameters set for a key generation session.
//
// Fields:
//   Algorithm (string): Algorithm name for keygen, must match OpenSSL
//   Debug     (bool): Enable verbose debug output (default: false)
//   Format    (KeyFormat): Output format for generated key
//   KeyPairMode (bool): Enable keypair mode (both PEM public & DER private)
type Config struct {
    Algorithm   string
    Debug       bool
    Format      KeyFormat // Empty if KeyPairMode=true
    KeyPairMode bool
}

// Load constructs a Config from environment variables, with validation.
//
// Returns:
//   (*Config): Populated with values from env (KEYGEN_ALGORITHM, DEBUG, KEYGEN_FORMAT, KEYGEN_KEYPAIR)
//   (error):  If required KEYGEN_ALGORITHM is unset or invalid
func Load() (*Config, error) {
    alg := os.Getenv("KEYGEN_ALGORITHM")
    if alg == "" {
        return nil, fmt.Errorf("environment variable KEYGEN_ALGORITHM is required")
    }
    debugEnv := os.Getenv("DEBUG")
    debug := debugEnv == "true"
    keypairEnv := os.Getenv("KEYGEN_KEYPAIR")
    keyPairMode := keypairEnv == "true"
    formatRaw := os.Getenv("KEYGEN_FORMAT")
    if keyPairMode && formatRaw != "" {
        return nil, fmt.Errorf("Cannot set both KEYGEN_KEYPAIR and KEYGEN_FORMAT. Use only one mode.")
    }
    var format KeyFormat
    if keyPairMode {
        format = "" // Ignore format completely
    } else {
        if formatRaw == "" {
            formats := GetSupportedKeyFormats()
            if len(formats) == 0 {
                return nil, fmt.Errorf("no key formats are registered in the system")
            }
            formatRaw = string(formats[0]) // First available format
        }
        formatRaw = strings.ToUpper(formatRaw)
        if !IsSupportedFormat(formatRaw) {
            return nil, fmt.Errorf("unsupported KEYGEN_FORMAT: %s (supported: %v)", formatRaw, GetSupportedKeyFormats())
        }
        for _, kf := range GetSupportedKeyFormats() {
            if strings.EqualFold(string(kf), formatRaw) {
                format = kf
                break
            }
        }
    }
    cfg := &Config{
        Algorithm:   alg,
        Debug:       debug,
        Format:      format,
        KeyPairMode: keyPairMode,
    }
    return cfg, nil
}
