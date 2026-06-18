#!/usr/bin/env bash

###############################################################################
# PQC Key Generator Orchestration Script
# Orchestrates build, run, and cleanup for post-quantum key generation in an
# isolated container
#
# Usage:
#   ./generate_key.sh [--key|-k <algorithm>] [--format|-f <format>] [--keypair|-kp]
#     --key/-k <string>     : (Optional) Set KEYGEN_ALGORITHM (default: ML-KEM-512)
#     --format/-f <string>  : (Optional) Set KEYGEN_FORMAT ('DER', 'PEM', ...) (default: DER)
#     --keypair/-kp         : (Optional) Enable keypair mode (outputs PEM+DER, disables --format/KEYGEN_FORMAT)
#
# Example:
#   ./generate_key.sh --key ML-KEM-1024 --keypair
#   ./generate_key.sh --key ML-KEM-1024 --format DER
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

show_help() {
    cat <<EOF
Usage: $0 [OPTIONS]

Post-quantum key generation in an isolated container using OpenSSL + OQS-provider.

Options:
  --podman, -p              Use podman instead of docker
  --docker, -d              Use docker instead of podman
  --foreground, -f          Run container in foreground
  --key, -k <algorithm>     Set key generation algorithm (overrides KEYGEN_ALGORITHM)
  --format, -f <format>     Set key file format (DER, PEM, ...), overrides KEYGEN_FORMAT
  --keypair, -kp            Enable keypair mode (outputs PEM public + DER private), disables --format and KEYGEN_FORMAT
  --help, -h                Show this help message

Examples:
  ./generate_key.sh --key ML-KEM-1024 --format PEM
  ./generate_key.sh --key ML-KEM-1024 --keypair
  CONTAINER_ENGINE=docker ./generate_key.sh -p
  Supported formats depend on your build/runtime and are set by KEYGEN_FORMAT or --format. Default: DER

EOF
}

# Absolute path to script directory
script_dir="$(cd "$(dirname "$0")" && pwd)"
cd "$script_dir"

ALG_ARG=""
FORMAT_ARG=""
KEYPAIR_MODE=""

# Parse command line arguments including the existing ones (--key, --format, --keypair)
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --podman|-p)
                CONTAINER_ENGINE="podman"
                shift
                ;;
            --docker|-d)
                CONTAINER_ENGINE="docker"
                shift
                ;;
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
            --keypair|-kp)
                KEYPAIR_MODE="true"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                warn "Unknown argument: $1"
                shift
                ;;
        esac
    done
}

# Read environment variables from file and return them as a space-separated string
read_env_file() {
    local env_file=""
    local vars_to_read=()
    local value_only=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --value|-v)
                value_only=true
                shift
                ;;
            --file|-f)
                if [[ -n "$2" && "$2" != "--"* ]]; then
                    env_file="$2"
                    shift 2
                else
                    error "Missing file argument for --file flag"
                fi
                ;;
            --file=*)
                env_file="${1#*=}"
                shift
                ;;
            *)
                if [[ -n "$env_file" ]]; then
                    IFS=',' read -ra vars <<< "$1"
                    for var in "${vars[@]}"; do
                        vars_to_read+=("$var")
                    done
                else
                    env_file="$1"
                fi
                shift
                ;;
        esac
    done
    
    if [[ -z "$env_file" ]]; then
        env_file=".env"
    fi
    
    if [[ ! -f "$env_file" ]]; then
        info "Environment file $env_file not found"
        return 1
    fi
    
    local env_args=""
    
    # Read and process each line
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        
        # Extract variable name and value
        local var_name="${line%%=*}"
        local var_value="${line#*=}"
        
        # Remove trailing comments and trim whitespace
        var_value="${var_value%%#*}"
        var_name="$(echo "$var_name" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        var_value="$(echo "$var_value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
        
        # Skip if variable name is empty
        [[ -z "$var_name" ]] && continue
        
        # If specific variables are requested, check if this one matches
        if [[ ${#vars_to_read[@]} -gt 0 ]]; then
            local found=false
            for var in "${vars_to_read[@]}"; do
                if [[ "$var" == "$var_name" ]]; then
                    found=true
                    break
                fi
            done
            [[ "$found" == false ]] && continue
        fi
        
        # Handle output format based on value_only flag
        if [[ "$value_only" == true ]]; then
            echo "$var_value"
        else
            # Return key=value pairs
            if [[ -n "$env_args" ]]; then
                env_args="$env_args $var_name=$var_value"
            else
                env_args="$var_name=$var_value"
            fi
        fi
    done < "$env_file"
    
    # Output all variables
    if [[ "$value_only" == false ]]; then
        echo "$env_args"
    fi
}

# Ensure all variables from the template are present in the actual env file
ensure_env_vars() {
    local template_file="$1"
    local env_file="$2"
    local updated=0
    info "Syncing .env with template: $template_file"
    
    # Read all variables from template
    local template_vars=$(read_env_file -f "$template_file")
    
    # Process each variable from the template
    for var in $template_vars; do
        [[ -z "$var" ]] && continue
        
        local var_name="${var%%=*}"
        [[ "$var_name" =~ ^[[:space:]]*# ]] && continue
        
        info "Checking if $var_name is present in $env_file ..."
        if ! grep -Eq "^[[:space:]]*#?[[:space:]]*$var_name[[:space:]]*=" "$env_file"; then
            last_char=$(tail -c1 "$env_file" 2>/dev/null || echo '')
            if [[ "$last_char" != "" && "$last_char" != $'\n' ]]; then
                echo >> "$env_file"
            fi
            echo "$var" >> "$env_file"
            info "Added $var_name to $env_file"
            updated=1
        fi
    done
    
    if [[ $updated -eq 1 ]]; then
        info "Completed variable sync: $env_file updated"
    else
        info "No missing variables detected in $env_file"
    fi
}

# Ensures .env exists, exports all config as env variables
make_env() {
    if [[ -f .env ]]; then
        info "Using existing .env"
        ensure_env_vars .env.example .env
    else
        [[ -f .env.example ]] || error "No key/.env.example template found"
        info "Creating .env from .env.example..."
        local temp_env=""
        temp_env=$(read_env_file -f .env.example)
        echo "$temp_env" | tr ' ' '\n' > .env
        success "Created .env from .env.example"
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
}

# Read and export environment variables into current context
set_env() {
    local env_args=$(read_env_file -f .env)
    
    # Export all variables to current context
    if [[ -n "$env_args" ]]; then
        export $env_args
    fi
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
    "$CONTAINER_ENGINE" build -t $IMAGE_NAME -f Containerfile "$script_dir" >/dev/null
}

# Run container, validate DER output, echo relative result path
run_keygen() {
    info "Running container..."
    local rel_key_path
    if ! rel_key_path=$("$CONTAINER_ENGINE" run --rm --env-file .env -v "$TMP:/mnt/key" $IMAGE_NAME 2>&1); then
        error "Container execution failed: $rel_key_path"
    fi
    
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
        echo "$abs_path1"
        echo "$abs_path2"
    else
        abs_path="$TMP/${rel_key_path##*/}"
        if [ ! -f "$abs_path" ]; then
            error "Key output file missing in container output: $abs_path (format=$KEYGEN_FORMAT)"
        fi
        echo "$abs_path"
    fi
}

# Main orchestration entrypoint
main() {
    parse_args "$@"
    make_env
    set_env
    detect_container_engine
    prepare_volume
    build_image
    run_keygen
    clean_ttl
}

main "$@"
