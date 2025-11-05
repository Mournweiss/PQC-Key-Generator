// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package formats

import (
    "strings"
    "sync"
    "pqckeygen/internal/formats/templates"
)

// registry.go implements a dynamic runtime registry for all key format workers.

var (
    registry      = make(map[string]templates.KeyFormatWorker)
    registryMutex sync.RWMutex
)

// RegisterFormatWorker registers a new format worker (must implement KeyFormatWorker).
// Should be called from the init() function of each format module.
//
// Params:
//   format (string): Unique format name (e.g., "DER", "PEM") — case-insensitive
//   worker (KeyFormatWorker): Handler for all keygen logic for this format
func RegisterFormatWorker(format string, worker templates.KeyFormatWorker) {
    registryMutex.Lock()
    defer registryMutex.Unlock()
    registry[strings.ToUpper(format)] = worker
}

// GetFormatWorker retrieves a registered worker by format name (case-insensitive).
// Returns the worker (if registered) or (nil, false) if not.
//
// Params:
//   format (string): Format to look up
//
// Returns:
//   (KeyFormatWorker, bool): The worker and its registry status
func GetFormatWorker(format string) (templates.KeyFormatWorker, bool) {
    registryMutex.RLock()
    defer registryMutex.RUnlock()
    worker, ok := registry[strings.ToUpper(format)]
    return worker, ok
}

// ListRegisteredFormats returns a slice with the names of all registered formats (as uppercase strings).
//
// Returns:
//   ([]string): All registered format type strings in the system (order not guaranteed)
func ListRegisteredFormats() []string {
    registryMutex.RLock()
    defer registryMutex.RUnlock()
    keys := make([]string, 0, len(registry))
    for k := range registry {
        keys = append(keys, k)
    }
    return keys
}
