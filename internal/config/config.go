package config

import (
    "fmt"
    "os"
)

// Config holds all environment-driven runtime parameters
//
// Fields:
//   Algorithm (string): Algorithm name for keygen, must match OpenSSL
//   Debug     (bool): Enable verbose debug output (default: false)
type Config struct {
    Algorithm string
    Debug     bool
}

// Load constructs a Config from environment variables, with validation
//
// Returns:
//   (*Config): Populated with values from env (KEYGEN_ALGORITHM, DEBUG)
//   (error):  If required KEYGEN_ALGORITHM is unset or invalid
func Load() (*Config, error) {
    alg := os.Getenv("KEYGEN_ALGORITHM")
    if alg == "" {
        return nil, fmt.Errorf("environment variable KEYGEN_ALGORITHM is required")
    }
    debugEnv := os.Getenv("DEBUG")
    debug := debugEnv == "true"
    cfg := &Config{
        Algorithm: alg,
        Debug:     debug,
    }
    return cfg, nil
}
