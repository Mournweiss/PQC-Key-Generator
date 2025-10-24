package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "log"
    "os"
    "pqckeygen/internal/config"
    pqc "pqckeygen/internal/pqc"
    validation "pqckeygen/internal/validation"
    errors "pqckeygen/internal/errors"
    "os/exec"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("%v", err)
    }

    if os.Getenv("DEBUG") == "true" {
        log.Println("Listing available public-key algorithms from OpenSSL:")
        listCmd := "openssl list -public-key-algorithms -provider oqsprovider"
        output, derr := exec.Command("sh", "-c", listCmd).CombinedOutput()
        if derr == nil {
            log.Printf("%s", output)
        } else {
            log.Printf("OpenSSL algorithm list failed: %v\n%s", derr, output)
        }
    }

    randomPart, err := randomHex(16)
    if err != nil {
        log.Fatalf("Could not generate secure random: %v", err)
    }
    outPath := "/mnt/key/" + randomPart + ".der"

    genPath, err := pqc.GenerateKey(cfg.Algorithm, outPath)
    if err != nil {
        logError("GENERATION ERROR:", err)
        os.Exit(1)
    }
    if vErr := validation.ValidateDERKey(genPath); vErr != nil {
        logError("VALIDATION ERROR:", vErr)
        os.Exit(1)
    }
    fmt.Println(genPath)
}

func randomHex(n int) (string, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}

func logError(prefix string, err error) {
    switch v := err.(type) {
    case *errors.KeyGenError:
        log.Printf("%s %s | Context: %+v\n", prefix, v.Message, v.Context)
    default:
        log.Printf("%s %v\n", prefix, err)
    }
}
