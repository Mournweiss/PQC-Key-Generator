package config

import (
    "fmt"
    "os"
)

type Config struct {
    Algorithm string
}

func Load() (*Config, error) {
    alg := os.Getenv("KEYGEN_ALGORITHM")
    if alg == "" {
        return nil, fmt.Errorf("environment variable KEYGEN_ALGORITHM is required")
    }
    cfg := &Config{
        Algorithm: alg,
    }
    return cfg, nil
}
