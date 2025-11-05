// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package templates

// KeyFormatWorker defines an interface for pluggable strategies for cryptographic key generation and validation per format.
//
// Each key format (e.g., DER, PEM) must provide an implementation that satisfies this interface.
//
// Methods:
//   Generate(algorithm, outPath): Generate a key with the given algorithm, writing to the given path. Returns the output path and any error.
//   Validate(outPath): Validate the key at the given path, returning error on invalid files.
type KeyFormatWorker interface {
    Generate(algorithm string, outPath string) (string, error)
    Validate(outPath string) error
}

// BaseFormatWorker is an embeddable struct for shared logic among format workers.
type BaseFormatWorker struct{}
