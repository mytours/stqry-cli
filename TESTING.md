# Testing

## Unit Tests

Run all unit tests:

```sh
make test
```

Or directly with Go:

```sh
go test ./...
```

Run tests for a specific package:

```sh
go test ./internal/api/...
go test ./internal/cli/...
```

Run a specific test by name:

```sh
go test ./internal/api/... -run TestListCollections
```

Run with verbose output:

```sh
go test ./... -v
```

## End-to-End Tests

E2e tests use [BATS](https://github.com/bats-core/bats-core) and exercise the compiled `stqry` binary.

### Prerequisites

```sh
brew install bats-core
```

### Running

```sh
make test-e2e
```

This will:
1. Build the `stqry` binary (`bin/stqry`)
2. Build the cassette replay proxy (`bin/recorder`)
3. Start the replay proxy if cassette files exist in `e2e/cassettes/happypath/`
4. Run all `*.bats` test files in `e2e/`

### Recording cassettes

The replay proxy intercepts HTTP calls made by the binary and plays back recorded responses. To record new cassettes, run the recorder binary in record mode against a real API, then commit the resulting JSON files under `e2e/cassettes/happypath/`.

### Running a single BATS file

```sh
make build build-recorder
e2e/run.sh e2e/config.bats
```

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `STQRY` | `bin/stqry` | Path to the binary under test |
| `REPLAY_PORT` | `8765` | Port used by the cassette replay proxy |

## Linting

```sh
make lint
```
