#!/usr/bin/env bash

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

script_dir="$(cd "$(dirname "$0")" && pwd)"
cd "$script_dir"

ALG_ARG=""
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
        *)
            shift
            ;;
    esac
done

make_env() {
    [[ -f .env ]] && { info "Using existing .env"; } || {
        [[ -f .env.example ]] || error "No key/.env.example template found"
        cp .env.example .env && success "Created default .env from .env.example"
    }
    if [[ -z "$ALG_ARG" ]]; then
        ALG_ARG="ML-KEM-512"
        info "No algorithm specified, defaulting to ML-KEM-512"
    fi
    if grep -q '^KEYGEN_ALGORITHM=' .env; then
        sed -i "s/^KEYGEN_ALGORITHM=.*/KEYGEN_ALGORITHM=$ALG_ARG/" .env
        info "Overriding KEYGEN_ALGORITHM in .env with: $ALG_ARG"
    else
        echo "KEYGEN_ALGORITHM=$ALG_ARG" >> .env
        info "Added KEYGEN_ALGORITHM to .env: $ALG_ARG"
    fi
    while IFS='=' read -r key value; do
        if [[ "$key" =~ ^[A-Z_][A-Z0-9_]*$ && -n "$value" ]]; then
            export "$key"="$value"
        fi
    done < <(grep -v '^#' .env | grep -v '^$')
}

resolve_tmp_dir() {
    local tmp
    if [[ -n "${TMP:-}" ]]; then
        tmp="${TMP// /}"
        [[ "$tmp" == /* ]] || tmp="$script_dir/$tmp"
    else
        tmp="$script_dir/keygen_tmp"
    fi
    printf "%s" "$tmp"
}

clean_ttl() {
    local ttl="${TMP_TTL_SEC:-600}"
    local tmp_dir="$(resolve_tmp_dir)"
    info "Setting timer to auto-clean TMP in $ttl seconds for $tmp_dir"
    nohup bash -c "sleep $ttl && rm -rf '$tmp_dir'" > /dev/null 2>&1 &
}

prepare_volume() {
    local tmp_dir="$(resolve_tmp_dir)"
    mkdir -p "$tmp_dir"
    export TMP="$tmp_dir"
    info "Preparing volume $TMP"
}

build_image() {
    info "Building keygen container..."
    podman build -t $IMAGE_NAME "$script_dir" >/dev/null
}

run_keygen() {
    info "Running container..."
    local rel_der_path
    rel_der_path=$(podman run --rm --env-file .env -v "$TMP:/mnt/key" $IMAGE_NAME)
    local out_name
    out_name="${rel_der_path#/mnt/key/}"
    local rel_out_path
    rel_out_path="$(basename "$TMP")/$out_name"
    local abs_path
    abs_path="$script_dir/$rel_out_path"
    if [ ! -f "$abs_path" ]; then
        error "Key DER file missing in container output: $abs_path"
    fi
    info "Key file ready: $rel_out_path"
    echo "$rel_out_path"
}

main() {
    make_env
    prepare_volume
    build_image
    local key_path
    key_path=$(run_keygen)
    abs_path="$script_dir/$key_path"
    echo "$abs_path"
    clean_ttl
}

main
