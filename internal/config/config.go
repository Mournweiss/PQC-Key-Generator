package config

import (
    "fmt"
    "os"
)

type Config struct {
    Algorithm string
    Debug     bool
}

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
