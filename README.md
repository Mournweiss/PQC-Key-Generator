<div align="center">

# PQC-Key-Generator

Containerized Key Generator

[![Authors](https://img.shields.io/badge/-AUTHORS-blue?style=for-the-badge&logoWidth=40)](AUTHORS.md)
[![License](https://img.shields.io/badge/License-MIT-green.svg?style=for-the-badge&logoWidth=40)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25.3-00ADD8?style=for-the-badge&logoWidth=40)](https://golang.org)
[![OpenSSL](https://img.shields.io/badge/OpenSSL-3.5.0-3C3C3D?style=for-the-badge&logoWidth=40)](https://www.openssl.org/)
[![liboqs](https://img.shields.io/badge/liboqs-main-blueviolet?style=for-the-badge&logoWidth=40)](https://github.com/open-quantum-safe/liboqs)
[![oqs-provider](https://img.shields.io/badge/OQS--Provider-main-purple?style=for-the-badge&logoWidth=40)](https://github.com/open-quantum-safe/oqs-provider)

</div>

## Overview

PQC-Key-Generator is a containerized toolkit for generating post-quantum cryptographic (PQC) keys with OpenSSL and OQS-provider.

Technology Stack:

-   **Go** (1.25.3)
-   **OpenSSL** (3.5.0) compiled with post-quantum algorithm support
-   **OQS-provider** (main branch)
-   **liboqs** (main branch)
-   **Podman/Docker** or compatible containerization engine

## Usage

1. Clone the repository:

    ```
    git clone https://github.com/Mournweiss/PQC-Key-Generator.git

    cd PQC-Key-Generator
    ```

2. Prepare and run [orchestration script](generate_key.sh):

    ```bash
    chmod +x generate_key.sh

    ./generate_key.sh --key ML-KEM-1024 --format DER
    ```

    ### Arguments:

    ```text
    --key, -k <algorithm>     Set key generation algorithm (overrides KEYGEN_ALGORITHM)
    --format, -f <format>     Set key file format (DER, PEM, ...), overrides KEYGEN_FORMAT
    --keypair, -p             Enable keypair mode (outputs PEM public + DER private); disables --format/KEYGEN_FORMAT
    --help, -h                Show this help message and exit

    Supported formats depend on your build/runtime and are set by KEYGEN_FORMAT or --format. Default: DER.
    ```

3. Get absolute path to key file:

    Script outputs the absolute path to the generated key file, for example:

    ```
    /home/user/PQC-Key-Generator/keygen_tmp/8bbbc7eedea23f0e4f23b4bf472fce20.der
    /home/user/PQC-Key-Generator/keygen_tmp/83ac534ff0e9286f1f8d524dcb3517a8.pem
    ```

    > Key file in the temporary directory will be automatically deleted after the TTL set by `TMP_TTL_SEC` (default: 5 seconds).

## Keypair Mode

**Key Pair Generation (PEM + DER) Mode**  
Enables atomic generation and output of both PEM (private, used for public output) and DER (private) files. This is triggered by:

```
./generate_key.sh --key ML-KEM-1024 --keypair
```

**In this mode, --format (KEYGEN_FORMAT) is disabled!**

### Output Example

```
/home/user/PQC-Key-Generator/keygen_tmp/abc1234.pem
/home/user/PQC-Key-Generator/keygen_tmp/abc1234.der
```

## Supported Key Formats

-   DER (ASN.1 binary: .der)
-   PEM (Privacy Enhanced Mail: .pem)

## Environment Variables

-   **KEYGEN_ALGORITHM**: The PQC algorithm for key generation. Must be supported by the linked OpenSSL build. (Default: `ML-KEM-512`) (See a list of supported algorithms in OQS-provider [here](https://github.com/open-quantum-safe/oqs-provider#algorithms)).

-   **KEYGEN_FORMAT**: Output format for generated key (DER, PEM, ...). Must match a supported format handler. Default: DER. Controls the format of the exported key file.

-   **KEYGEN_KEYPAIR**: `true` to enable pair mode, disables KEYGEN_FORMAT

-   **DEBUG**: Enable verbose OpenSSL debug output (`true` or `false`). Helpful for troubleshooting algorithm/provider issues. (Default: `false`)

-   **IMAGE_NAME**: The container image name used for key generation. Customize to avoid conflicts in your environment. (Default: `pqckeygen`)

-   **TMP**: Directory for temporary key output (mapped as container volume). Ensure it is writable and persistent for duration of operation. (Default: `keygen_tmp`)

-   **TMP_TTL_SEC**: Time (in seconds) after which the temporary directory and its contents are auto-cleaned up by the orchestration script. Increase for debugging or persistent storage. (Default: `5`)

-   **CONTAINER_ENGINE**: Override backend autodetection; set to `docker` or `podman` to specify, else leave empty for automatic selection.
