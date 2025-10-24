package pqc

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    errors "pqckeygen/internal/errors"
    utils "pqckeygen/internal/utils"
)

// GenerateKey generates a cryptographic keypair as DER at outPath using the supplied algorithm.
//
// Parameters:
//   algorithm string: The algorithm name as required by OpenSSL (case-insensitive match)
//   outPath   string: Target DER file path (volume-mapped, will also write <outPath>.pem temporarily)
//
// Returns:
//   (string)  - Path to successful DER key file (same as outPath)
//   (error)   - If generation fails due to OpenSSL or unsupported algorithm
func GenerateKey(algorithm string, outPath string) (string, error) {
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
            Message: fmt.Sprintf("%s not found in OpenSSL list", algorithm),
            Context: map[string]interface{}{ "output": string(out), "algorithm": algorithm },
        }}
    }

    pemFile := outPath + ".pem"
    genCmd := exec.Command("openssl", "genpkey", "-provider", "default", "-algorithm", algorithm, "-out", pemFile)
    genOut, err := genCmd.CombinedOutput()
    if err != nil || !utils.FileExists(pemFile) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("Key generation via openssl failed for %s", algorithm),
            Context: map[string]interface{}{ "cmd": genCmd.String(), "output": string(genOut), "path": pemFile, "algorithm": algorithm, "log": string(genOut) },
        }}
    }

    derCmd := exec.Command("openssl", "pkey", "-in", pemFile, "-outform", "DER", "-out", outPath)
    derOut, err := derCmd.CombinedOutput()
    if err != nil || !utils.FileExists(outPath) {
        return "", &errors.PQCNotSupportedError{errors.KeyGenError{
            Message: fmt.Sprintf("DER export failed for %s", algorithm),
            Context: map[string]interface{}{ "cmd": derCmd.String(), "output": string(derOut), "path": outPath, "algorithm": algorithm, "log": string(derOut) },
        }}
    }
    _ = os.Remove(pemFile)

    return outPath, nil
}
