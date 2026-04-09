#!/usr/bin/env bash
# Run e2e tests against the stqry binary using BATS.
# Optionally starts the cassette replay proxy if cassettes and the recorder binary exist.
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Load .env from repo root if present
if [ -f "$ROOT_DIR/.env" ]; then
    set -a
    # shellcheck source=/dev/null
    source "$ROOT_DIR/.env"
    set +a
fi

export REPLAY_PORT="${REPLAY_PORT:-8765}"
export STQRY="${STQRY:-${ROOT_DIR}/bin/stqry}"

# Check prerequisites
if [ ! -f "$STQRY" ]; then
    echo "Error: stqry binary not found at $STQRY. Run 'make build' first."
    exit 1
fi

if ! command -v bats &>/dev/null; then
    echo "Error: bats not found. Install with: brew install bats-core"
    exit 1
fi

# Start replay proxy if cassettes and the recorder binary are available
CASSETTES_DIR="${SCRIPT_DIR}/cassettes/happypath"
RECORDER="${ROOT_DIR}/bin/recorder"

RECORDER_PID=""

if [ -d "$CASSETTES_DIR" ] && [ -f "$RECORDER" ]; then
    # Only start the proxy if there are actual cassette files
    if ls "$CASSETTES_DIR"/*.json &>/dev/null 2>&1; then
        echo "Starting replay proxy on port $REPLAY_PORT..."
        "$RECORDER" --mode=replay --port="$REPLAY_PORT" --cassettes="$CASSETTES_DIR" &
        RECORDER_PID=$!
        # Give the server a moment to bind
        sleep 0.5
    else
        echo "No cassette files found in $CASSETTES_DIR, skipping replay proxy."
    fi
fi

cleanup() {
    if [ -n "$RECORDER_PID" ]; then
        kill "$RECORDER_PID" 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Run all bats test files in the e2e directory
exec bats "${SCRIPT_DIR}"/*.bats "$@"
