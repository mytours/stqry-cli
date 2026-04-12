#!/usr/bin/env bash
# Common setup/teardown and assertion helpers for BATS tests

setup() {
    # Create isolated home and work directories
    export TEST_HOME
    TEST_HOME=$(mktemp -d)
    export HOME="$TEST_HOME"
    export STQRY_CONFIG_HOME="$TEST_HOME/.config/stqry"

    # Create temp working dir
    export TEST_WORK
    TEST_WORK=$(mktemp -d)
    cd "$TEST_WORK" || return 1

    # Clear any existing env vars
    unset STQRY_TOKEN
    unset STQRY_API_URL
}

teardown() {
    rm -rf "$TEST_HOME" "$TEST_WORK"
}

# Create global config with a test site.
# Defaults: TEST_API_URL env var → replay proxy; TEST_TOKEN env var → "test-token"
create_global_config() {
    local api_url="${1:-${TEST_API_URL:-http://localhost:${REPLAY_PORT:-8765}}}"
    local token="${2:-${TEST_TOKEN:-test-token}}"
    mkdir -p "$TEST_HOME/.config/stqry"
    cat > "$TEST_HOME/.config/stqry/config.yaml" <<EOF
sites:
  testsite:
    token: $token
    api_url: $api_url
EOF
}

# Create a directory-level stqry.yaml in the current working directory.
# Defaults to site=testsite so doctor's directory config check passes.
create_directory_config() {
    local site="${1:-testsite}"
    cat > stqry.yaml <<EOF
site: $site
EOF
}

# Assertion helpers

assert_success() {
    if [ "$status" -ne 0 ]; then
        echo "Expected success (exit 0), got: $status"
        echo "Output: $output"
        return 1
    fi
}

assert_failure() {
    if [ "$status" -eq 0 ]; then
        echo "Expected failure (non-zero exit), but got: $status"
        echo "Output: $output"
        return 1
    fi
}

assert_output_contains() {
    if ! echo "$output" | grep -qF -- "$1"; then
        echo "Expected output to contain: $1"
        echo "Actual output: $output"
        return 1
    fi
}

assert_json_value() {
    local key="$1"
    local expected="$2"
    local actual
    actual=$(echo "$output" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d$key)" 2>/dev/null)
    if [ "$actual" != "$expected" ]; then
        echo "Expected $key = $expected, got: $actual"
        return 1
    fi
}

# Find the stqry binary — prefer explicit STQRY env var, otherwise look next to this file.
STQRY="${STQRY:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/bin/stqry}"
