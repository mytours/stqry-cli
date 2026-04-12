# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build binary with version from git tags → bin/stqry
make test           # Run unit tests (go test ./... -v)
make lint           # go vet ./...
make test-e2e       # Build binary + recorder, run BATS integration tests
make clean          # Remove bin/
```

**Run a single unit test:**
```bash
go test ./internal/cli/... -run TestCollectionsListCmd -v
go test ./internal/api/... -run TestListCollections -v
```

**Run a single BATS e2e file:**
```bash
make build build-recorder
e2e/run.sh e2e/config.bats
```

**Record new API cassettes** (requires `TEST_API_URL` in `.env`):
```bash
make record TEST_API_URL=https://api-us.stqry.com
```

## Architecture

Entry point: `cmd/stqry/main.go` → `internal/cli.Execute()`

### Layer structure

**`internal/cli/`** — Cobra command factories (`newXxxCmd()`)
- Each command is a closure using package-level vars: `activeClient`, `printer`, `globalConfig`, `flagSite`, `flagJSON`, `flagQuiet`
- `root.go`'s `PersistentPreRunE` initializes these globals; skipped for commands in `skipSiteResolution()` (config, setup, doctor, mcp, help, completion)
- Commands call into `internal/api/` and pass results to `printer`

**`internal/api/`** — HTTP client
- `Client` wraps `net/http` with 30s timeout, sends `X-Api-Token` header
- Methods: `Get()`, `Post()`, `Patch()`, `Put()`, `Delete()` — all return `(map[string]interface{}, error)`
- Per-resource files: `collections.go`, `screens.go`, `media.go`, etc.
- List methods return `([]map[string]interface{}, PaginationMeta, error)`
- Custom `APIError` type with `.Code` and `.Message`

**`internal/output/`** — Three output modes
- Human: text/tabwriter tables
- JSON (`--json`): `{data: [...], meta: {...}}` envelope
- Quiet (`--quiet`): raw data array, no envelope — intended for piping to `jq`

**`internal/config/`** — Site resolution (YAML)
- Global: `~/.config/stqry/config.yaml`
- Local: `stqry.yaml` or `stqry.yml` (walks directory tree upward)
- Resolution: `--site` flag > local file > global config

**`internal/mcp/`** — MCP server over stdio (mark3labs/mcp-go)
- Mirrors CLI commands as MCP tools; session-level in-memory site selection via `connect`/`select_site` tools
- Doctor and config logic must be kept in sync between CLI and MCP implementations

**`internal/doctor/`** — Health checks shared between `stqry doctor` and the MCP server

### Adding a new command

1. Add `newXxxCmd()` in `internal/cli/xxx.go` using `cobra.Command` with `RunE`
2. Implement API methods in `internal/api/xxx.go` using `Client.Get/Post/Patch/Put/Delete`
3. Add unit tests in `internal/cli/xxx_test.go` with `httptest.NewServer` and `setupTestHome()`
4. If the command manages STQRY content (not a meta/tooling command like `setup` or `mcp`), add tool registration in `internal/mcp/tools_xxx.go` so AI agents can use it

### Testing patterns

Unit tests use:
- `setupTestHome(t, configYAML)` — creates temp HOME directory with a config pointing to a mock server
- `httptest.NewServer(...)` — intercepts API calls
- `os.Pipe()` — captures stdout (printer writes directly to `os.Stdout`)
- `newRootCmd().SetArgs([]string{...})` — runs commands in-process

E2E tests use BATS + a cassette replay proxy. Cassettes are pre-recorded HTTP responses in `e2e/cassettes/happypath/`.

### Validation constants

Screen types: `story`, `web`, `panorama`, `ar`, `kiosk`
Media types: `map`, `webpackage`, `animation`, `audio`, `image`, `video`, `webvideo`, `ar`, `data`
Collection types: see `internal/cli/collections.go`
Region→URL mapping is centralized in `internal/api/` — add new regions there.
