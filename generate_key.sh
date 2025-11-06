#!/usr/bin/env bash

###############################################################################
# PQC Key Generator Orchestration Script
# Orchestrates build, run, and cleanup for post-quantum key generation in an
# isolated container
#
# Usage:
#   ./generate_key.sh [--key|-k <algorithm>] [--format|-f <format>] [--keypair|-p]
#     --key/-k <string>    : (Optional) Set KEYGEN_ALGORITHM (default: ML-KEM-512)
#     --format/-f <string> : (Optional) Set KEYGEN_FORMAT ('DER', 'PEM', ...) (default: DER)
#     --keypair/-p         : (Optional) Enable keypair mode (outputs PEM+DER, disables --format/KEYGEN_FORMAT)
#
# Example:
#   ./generate_key.sh --key ML-KEM-1024 --keypair
#   ./generate_key.sh --key ML-KEM-1024 --format DER
#
# Flow:
#   - Parse CLI for algorithm, format, and keypair mode
#   - Ensure .env, override KEYGEN_ALGORITHM/KEYGEN_FORMAT/KEYGEN_KEYPAIR if flags given
#   - Export all config from .env
#   - Prepare temp dir/volume for result
#   - Build the container image if needed
#   - Run generation in the container, validate output, print result path
#   - Schedule temp Dir cleanup
#
# Returns:
#   Absolute path to generated key(s),
#   or exits non-zero on error
###############################################################################

set -euo pipefail

# ANSI color codes
COLOR_INFO="\033[0m"       # White (default)
COLOR_WARN="\033[1;33m"    # Yellow
COLOR_ERROR="\033[1;31m"   # Red
COLOR_SUCCESS="\033[1;32m" # Green
COLOR_RESET="\033[0m"

info()    { echo -e "${COLOR_INFO}$1${COLOR_RESET}" >&2; }
warn()    { echo -e "${COLOR_WARN}$1${COLOR_RESET}" >&2; }
error()   { echo -e "${COLOR_ERROR}$1${COLOR_RESET}" >&2; exit 1; }
success() { echo -e "${COLOR_SUCCESS}$1${COLOR_RESET}" >&2; }

# Absolute path to script directory
script_dir="$(cd "$(dirname "$0")" && pwd)"
cd "$script_dir"

# Vars for CLI parsing
ALG_ARG=""
FORMAT_ARG=""
KEYPAIR_MODE=""

show_help() {
    cat <<EOF
Usage: ./generate_key.sh [OPTIONS]

Post-quantum key generation in an isolated container using OpenSSL + OQS-provider.

Options:
  --key, -k <algorithm>     Set key generation algorithm (overrides KEYGEN_ALGORITHM)
  --format, -f <format>     Set key file format (DER, PEM, ...), overrides KEYGEN_FORMAT
  --keypair, -p             Enable keypair mode (outputs PEM public + DER private), disables --format and KEYGEN_FORMAT
  --help, -h                Show this help message and exit

Examples:
  ./generate_key.sh --key ML-KEM-1024 --format PEM
  ./generate_key.sh --key ML-KEM-1024 --keypair
  CONTAINER_ENGINE=docker ./generate_key.sh -p
  Supported formats depend on your build/runtime and are set by KEYGEN_FORMAT or --format. Default: DER
EOF
}

# Ensure all variables from the template are present in the actual env file
ensure_env_vars() {
    local template_file="$1"
    local env_file="$2"
    local updated=0
    info "Syncing .env with template: $template_file"
    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" =~ ^# ]] && continue
        # Parse and trim var name
        var_name="${line%%=*}"
        var_name="$(echo "$var_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        info "Checking if $var_name is present in $env_file ..."
        if ! grep -Eq "^[[:space:]]*#?[[:space:]]*$var_name[[:space:]]*=" "$env_file"; then
            last_char=$(tail -c1 "$env_file" 2>/dev/null || echo '')
            if [[ "$last_char" != "" && "$last_char" != $'\n' ]]; then
                echo >> "$env_file"
            fi
            echo "$line" >> "$env_file"
            info "Added $var_name to $env_file"
            updated=1
        fi
    done < "$template_file"
    if [[ $updated -eq 1 ]]; then
        info "Completed variable sync: $env_file updated"
    else
        info "No missing variables detected in $env_file"
    fi
}

# Ensures .env exists, updates KEYGEN_ALGORITHM & KEYGEN_FORMAT, exports all config as env variables
make_env() {
    if [[ -f .env ]]; then
        info "Using existing .env"
        ensure_env_vars .env.example .env
    else
        [[ -f .env.example ]] || error "No key/.env.example template found"
        cp .env.example .env && success "Created default .env from .env.example"
    fi
    local session_ts="$(date +%s%N)_$RANDOM"
    export TMP="$PWD/keygen_tmp/session_$session_ts"
    info "Session unique TMP dir: $TMP"
    if [[ -n "$ALG_ARG" ]]; then
        sed -i "s/^KEYGEN_ALGORITHM=.*/KEYGEN_ALGORITHM=$ALG_ARG/" .env
        info "Overriding KEYGEN_ALGORITHM in .env with: $ALG_ARG"
    fi
    if [[ -n "$KEYPAIR_MODE" ]]; then
        sed -i "/^KEYGEN_FORMAT=/d" .env
        sed -i "s/^KEYGEN_KEYPAIR=.*/KEYGEN_KEYPAIR=true/" .env || echo "KEYGEN_KEYPAIR=true" >> .env
        info "Enabled keypair mode (KEYGEN_KEYPAIR=true), disabling individual format settings."
    else
        sed -i "/^KEYGEN_KEYPAIR=/d" .env
        if [[ -n "$FORMAT_ARG" ]]; then
            sed -i "s/^KEYGEN_FORMAT=.*/KEYGEN_FORMAT=$FORMAT_ARG/" .env || echo "KEYGEN_FORMAT=$FORMAT_ARG" >> .env
            info "Overriding KEYGEN_FORMAT in .env with: $FORMAT_ARG"
        fi
    fi
    while IFS='=' read -r key value; do
        if [[ "$key" =~ ^[A-Z_][A-Z0-9_]*$ && -n "$value" ]]; then
            export "$key"="$value"
        fi
    done < <(grep -v '^#' .env | grep -v '^$')
}

# Detect best available containerization engine (docker/podman) or use env override
detect_container_engine() {
    if [[ -n "${CONTAINER_ENGINE:-}" ]]; then
        info "Using user-specified containerization engine: $CONTAINER_ENGINE"
    elif command -v docker &>/dev/null; then
        CONTAINER_ENGINE=docker
        info "Using docker as a containerization engine"
    elif command -v podman &>/dev/null; then
        CONTAINER_ENGINE=podman
        info "Using podman as a containerization engine"
    else
        error "Neither Docker nor Podman found in \$PATH. Please install one or set CONTAINER_ENGINE."
    fi
}


# Returns path to temp output dir for key file volume mount
resolve_tmp_dir() {
    printf "%s" "$TMP"
}

# Launch background job to clean temp dir on TTL expiry
clean_ttl() {
    local ttl="${TMP_TTL_SEC:-600}"
    local tmp_dir="$TMP"
    info "Setting timer to auto-clean TMP in $ttl seconds for $tmp_dir"
    nohup bash -c "sleep $ttl && rm -rf '$tmp_dir'" > /dev/null 2>&1 &
}

# Create and export dir for output file container volume
prepare_volume() {
    local tmp_dir="$TMP"
    mkdir -p "$tmp_dir"
    export TMP="$tmp_dir"
    info "Preparing volume $TMP"
}

# Build container image for keygen service
build_image() {
    "$CONTAINER_ENGINE" build -t $IMAGE_NAME "$script_dir" >/dev/null
}

# Run container, validate DER output, echo relative result path
run_keygen() {
    info "Running container..."
    local rel_key_path
    rel_key_path=$("$CONTAINER_ENGINE" run --rm --env-file .env -v "$TMP:/mnt/key" $IMAGE_NAME)
    if [[ "$rel_key_path" == *,* ]]; then
        IFS="," read -r file1 file2 <<< "$rel_key_path"
        local abs_path1 abs_path2
        abs_path1="$TMP/${file1##*/}"
        abs_path2="$TMP/${file2##*/}"
        for abs in "$abs_path1" "$abs_path2"; do
            if [ ! -f "$abs" ]; then
                error "Key output file missing in container output: $abs (keypair mode)"
            fi
        done
        info "Key files ready: $abs_path1 and $abs_path2 (keypair mode)"
        echo "$abs_path1"
        echo "$abs_path2"
    else
        abs_path="$TMP/${rel_key_path##*/}"
        if [ ! -f "$abs_path" ]; then
            error "Key output file missing in container output: $abs_path (format=$KEYGEN_FORMAT)"
        fi
        info "Key file ready: $abs_path (format=$KEYGEN_FORMAT)"
        echo "$abs_path"
    fi
}

# Main orchestration entrypoint
main() {
    make_env
    detect_container_engine
    prepare_volume
    build_image
    run_keygen
    clean_ttl
}

# CLI argument parsing (support --format/-f and --keypair/-p)
while [[ $# -gt 0 ]]; do
    case $1 in
        --key|-k)
            if [[ -n "${2:-}" ]]; then
                ALG_ARG="$2"
                shift 2
            else
                error "Option $1 requires an argument (algorithm name)"
            fi
            ;;
        --format|-f)
            if [[ -n "${2:-}" ]]; then
                FORMAT_ARG="$2"
                shift 2
            else
                error "Option $1 requires an argument (format: DER or PEM)"
            fi
            ;;
        --keypair|-p)
            KEYPAIR_MODE="true"
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            shift
            ;;
    esac
done

main
