# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Releasing a new version

1. Update `CHANGELOG.md` — move items from `[Unreleased]` into a new versioned section (e.g. `## [0.6.0] - YYYY-MM-DD`), grouped under `### Added`, `### Fixed`, `### Removed`, etc.
2. Commit the changelog update.
3. Create and push a git tag: `git tag v0.6.0 && git push origin v0.6.0`
4. GitHub Actions picks up the tag and runs the release workflow: tests → GoReleaser (binaries + Homebrew tap) → PyPI publish.

Determine the version bump from commits since the last tag: new features → minor bump; fixes/docs only → patch bump.

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

**`internal/agentsmd/`** — Embeds `AGENTS.md` for writing to the CWD during `stqry config init`
- `internal/agentsmd/AGENTS.md` mirrors the root `AGENTS.md` — update both when the content changes

**`internal/mcp/`** — MCP server over stdio (mark3labs/mcp-go)
- Mirrors CLI commands as MCP tools; session-level in-memory site selection via `connect`/`select_site` tools
- Doctor and config logic must be kept in sync between CLI and MCP implementations

**`internal/doctor/`** — Health checks shared between `stqry doctor` and the MCP server

### Adding a new command

1. Add `newXxxCmd()` in `internal/cli/xxx.go` using `cobra.Command` with `RunE`
2. Implement API methods in `internal/api/xxx.go` using `Client.Get/Post/Patch/Put/Delete`
3. Add unit tests in `internal/cli/xxx_test.go` with `httptest.NewServer` and `setupTestHome()`
4. If the command manages STQRY content (not a meta/tooling command like `setup` or `mcp`), add tool registration in `internal/mcp/tools_xxx.go` so AI agents can use it
5. If the command maps to a public API endpoint in `docs/public_api.json`, update `API-COVERAGE.md`: add the CLI command to the endpoint row, change status to ✅, update the coverage count in the header, and remove the entry from Future Work if present

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
