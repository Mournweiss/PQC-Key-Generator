<div align="center">

# PQC-Key-Generator

Containerized Key Generator

[![Authors](https://img.shields.io/badge/-AUTHORS-blue)](AUTHORS.md)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-brightgreen?logo=go&logoColor=white)](https://golang.org)
[![OpenSSL](https://img.shields.io/badge/OpenSSL-3.5+-informational?logo=openssl&logoColor=white)](https://www.openssl.org/)
[![liboqs](https://img.shields.io/badge/liboqs-main-blueviolet)](https://github.com/open-quantum-safe/liboqs)
[![oqs-provider](https://img.shields.io/badge/OQS--Provider-main-purple)](https://github.com/open-quantum-safe/oqs-provider)

</div>

## Overview

PQC-Key-Generator is a containerized toolkit for generating post-quantum cryptographic (PQC) keys with OpenSSL and OQS-provider.

Technology Stack:

-   **Go** (1.25.3)
-   **OpenSSL** (3.5.0) compiled with post-quantum algorithm support
-   **OQS-provider** (main branch)
-   **liboqs** (main branch)
-   **Podman** or compatible container engine

## Usage

1. Clone the repository:

    ```
    git clone https://github.com/Mournweiss/PQC-Key-Generator.git

    cd PQC-Key-Generator
    ```

1. Prepare and run [orchestration script](generate_key.sh):

    ```bash
    chmod +x generate_key.sh

    ./generate_key.sh --key ML-KEM-1024
    ```

    This script will:

    - Setup and export env variables, update algorithm if `--key|-k` is passed
    - Prepare temp volumes for output
    - Build and run a container with selected algorithm
    - Output the absolute path to the generated DER key

    > By default, if no algorithm is specified via the CLI `--key` option or in the environment, the generator uses `ML-KEM-512`.

1. Get absolute path to key file:

    Script outputs the absolute path to the generated DER file, for example:

    ```
    /home/user/Desktop/PQC-Key-Generator/keygen_tmp/8bbbc7eedea23f0e4f23b4bf472fce20.der
    ```

    > Key file in the temporary directory will be automatically deleted after the TTL set by `TMP_TTL_SEC` (default: 5 seconds).

## Environment Variables

-   **KEYGEN_ALGORITHM**: The PQC algorithm for key generation. Must be supported by the linked OpenSSL build. (Default: `ML-KEM-512`)

-   **DEBUG**: Enable verbose OpenSSL debug output (`true` or `false`). Helpful for troubleshooting algorithm/provider issues. (Default: `false`)

-   **IMAGE_NAME**: The container image name used for key generation. Customize to avoid conflicts in your environment. (Default: `pqckeygen`)

-   **TMP**: Directory for temporary key output (mapped as container volume). Ensure it is writable and persistent for duration of operation. (Default: `keygen_tmp`)

-   **TMP_TTL_SEC**: Time (in seconds) after which the temporary directory and its contents are auto-cleaned up by the orchestration script. Increase for debugging or persistent storage. (Default: `5`)
