// SPDX-FileCopyrightText: 2025 Maxim Selin <selinmax05@mail.ru>
//
// SPDX-License-Identifier: MIT

package main

import (
    "fmt"
    "log"
    "os"
    "pqckeygen/internal/config"
    pqc "pqckeygen/internal/pqc"
    utils "pqckeygen/internal/utils"
    "os/exec"
)

// Entry point for the keygen CLI.
//
// Loads configuration, invokes format-agnostic key generation, and validates the result.
//
// Returns:
//   (os.Exit) with 0 on success, 1 on error (fatal/logged to stderr)
func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("[FATAL] %v", err)
    }

    if cfg.Debug {
        log.Println("Listing available public-key algorithms from OpenSSL:")
        listCmd := "openssl list -public-key-algorithms -provider default"
        output, derr := exec.Command("sh", "-c", listCmd).CombinedOutput()
        if derr == nil {
            log.Printf("%s", output)
        } else {
            log.Printf("OpenSSL algorithm list failed: %v\n%s", derr, output)
        }
    }

    randomPart, err := utils.RandomHex(16)
    if err != nil {
        log.Fatalf("Could not generate secure random: %v", err)
    }

    baseOutPath := "/mnt/key/" + randomPart
    genResult, err := pqc.GenerateKey(cfg, baseOutPath)
    if err != nil {
        log.Printf("GENERATION ERROR: %v\n", err)
        os.Exit(1)
    }
    if cfg.KeyPairMode {
        results := genResult
        var pemPath, derPath string
        n, _ := fmt.Sscanf(results, "%[^,],%s", &pemPath, &derPath)
        if n == 2 {
            fmt.Println(pemPath)
            fmt.Println(derPath)
        } else {
            fmt.Printf("%s\n", results)
        }
    } else {
        fmt.Println(genResult)
    }
}
