# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.10.3] - 2026-04-15

### Added

- `--short-title` flag on `stqry collections create|update` and `stqry screens create|update`. The API requires `short_title`, so create commands previously failed with `Short title can't be blank` whenever `--title` was passed. When `--short-title` is omitted on create, it defaults to `--title`.
- `--body`, `--subtitle`, `--description`, `--media-item-id`, and `--text-position` flags on `stqry screens sections add|update` so rich story content (paragraph text, image captions, single-media attachments) can be authored end-to-end from the CLI. `--text-position` defaults to `none` for `single_media` sections, which the API otherwise rejects with `Text position is not included in the list`.
- `--cover-image-media-item-id`, `--cover-image-grid-media-item-id`, `--cover-image-wide-media-item-id`, `--background-image-media-item-id`, and `--logo-media-item-id` flags on `stqry screens update` so screen header artwork can be set from the CLI.
- `--description`, `--cover-image-media-item-id`, `--cover-image-grid-media-item-id`, `--cover-image-wide-media-item-id`, `--logo-media-item-id`, `--preview-media-item-id`, and `--map-view-enabled` flags on `stqry collections update` so collection metadata, artwork, and the map view can be set from the CLI.
- `--lat`, `--lng`, `--map-type`, `--display-address`, and `--directions-enabled` flags on `stqry screens sections add` for building `location` sections. `--map-type` defaults to `single_location` when `--lat`/`--lng` are set.

## [0.10.2] - 2026-04-15

### Fixed

- `stqry screens update --title` and `stqry screens sections add|update --title` now correctly send a translated map (e.g. `{"en": "..."}`) when `--lang` is omitted. Previously they sent a plain string, which the API rejected with a 500 — making the section commands unusable without remembering to add `--lang en`.

## [0.10.1] - 2026-04-15

### Added

- `--short-title` flag on `stqry collections create|update` and `stqry screens create|update`. The API requires `short_title`, so create commands previously failed with `Short title can't be blank` whenever `--title` was passed. When `--short-title` is omitted on create, it defaults to `--title`.

## [0.10.0] - 2026-04-15

### Added

- `stqry config init` now writes `AGENTS.md` to the current working directory, giving AI agents immediate project context

### Fixed

- `stqry media upload` now requires `--media-id`; running without it would orphan the uploaded file (invisible in STQRY Builder). Use `stqry media create` to create a new media item with a file.
- Skills reference: teach Claude that resource subcommands are always plural (`screens`, `collections`, …)

### Changed

- Human-readable output now renders nested maps and translated fields as indented blocks instead of raw `map[string]interface{}` text; scalar slices render as comma-separated values

## [0.9.2] - 2026-04-15

- Version bump.

## [0.9.0] - 2026-04-15

### Added

- `stqry config show` — prints the resolved site config with source tracking (flag / local file / global)
- `stqry config validate` — checks the resolved config and performs a live token verification against the API
- CI workflow now uploads `stqry-skill.zip` to Bunny CDN on every tagged release, providing a stable download URL at `stqry-download/stqry-cli/stqry-skill.zip`
- `API-COVERAGE.md` — maps all 66 public API operations to their CLI commands (85% coverage)
- Expanded unit test coverage across codes, media, projects, and screens CLI commands; added E2E BATS stubs and `test-race`/`test-coverage` Makefile targets

## [0.8.0] - 2026-04-14

### Added

- `stqry skill export` command — packages skills as a Claude Desktop zip file (`stqry-skill.zip`)
- CI workflow uploads `stqry-skill.zip` as a GitHub release asset on every tagged release

### Changed

- `stqry setup claude --desktop` is deprecated; users are directed to `stqry skill export` instead

## [0.7.0] - 2026-04-14

### Added

- Python package (`stqry` on PyPI) now ships platform-specific wheels containing the Go binary — `pip install stqry` works on macOS, Linux, and Windows without a separate install step
- `build_wheels.py` script downloads GoReleaser archives and assembles wheels per platform
- Binary runner wrapper (`_run.py`) so the `stqry` console script delegates to the bundled binary

### Changed

- Replaced the previous pure-Python SDK with a thin binary wrapper; the Go CLI is now the sole implementation

## [0.6.3] - 2026-04-13

### Fixed

- Doctor: fixed false "Update available" warning when CLI version matches latest release (v-prefix mismatch)

## [0.6.2] - 2026-04-13

### Fixed

- Skills reference: taught Claude to prefer built-in `--jq` flag for filtering output

## [0.6.1] - 2026-04-13

### Fixed

- Skill export: merged duplicate YAML frontmatter blocks into one so `name`/`description` and version fields appear in a single `---` header

## [0.6.0] - 2026-04-13

### Added

- `stqry skill dump` command — prints the raw content of an installed skill file
- Skills versioning and update awareness: hash, frontmatter, and name helpers for installed skills
- Skills staleness checks wired into `stqry doctor` — warns when installed skills are out of date
- `--desktop` flag on `stqry setup` — exports skills as files for manual Claude Desktop installation
- `add_global_site` MCP tool — saves credentials to global config from within an AI session
- `site_name` parameter on `configure_project` MCP tool — writes a named-site local config
- MCP `connect` and `select_site` tools now suggest saving config when no `stqry.yaml` exists

### Fixed

- Skills reference: corrected screens sections commands and sub-item flags throughout
- Doctor: polished skills check messages and guarded against empty working directory
- Doctor: skip skills check when install directory is unavailable on the current platform
- MCP `configure_project`: session now set correctly when `site_name` is provided
- MCP `add_global_site`: improved error message on failure

### Removed

- Desktop-specific skill layout code (`LayoutDesktop`, `DesktopSkillsDir`) replaced by path-agnostic install logic

## [0.5.0] - 2026-04-13

### Added

- Shell completions for bash, zsh, fish, and PowerShell (`stqry completion <shell>`)
- Per-site completion cache with 1-hour TTL
- `completion status` subcommand — shows cache age and item counts
- `completion refresh` subcommand — repopulates the cache via paginated API calls
- Tab-completion for collection, screen, media, and project IDs via `ValidArgsFunction`
- `--jq` global flag with built-in jq filtering (powered by `gojq`)
- Rich usage examples added to all command help text
- `AGENTS.md` and updated `CLAUDE.md` for AI agent integration documentation

## [0.4.0] - 2026-04-12

### Added

- `stqry doctor` command — diagnostics for config, API connectivity, and version currency
  - Checks config file presence and validity
  - Checks API reachability and authentication
  - Checks whether the CLI is up to date (compares against the latest GitHub release)
  - `--verbose` flag for detailed per-check output; machine-readable `--output json`
- `run_doctor` MCP tool — exposes the same diagnostics to AI agents
- `internal/buildinfo` package — version variable injected at build time via `-ldflags`

## [0.3.1] - 2026-04-12

### Added

- `connect(token, api_url)` MCP tool — stores credentials in-memory for the session; no disk writes required
- `create_media` MCP tool — uploads a file and creates a new media item
- `Session.Clear()` for future deauthentication support
- Media type validation in `create_media` (rejects unknown types with a helpful error)
- Absolute path enforcement for `create_media` `file_path` parameter

### Fixed

- `select_site` and `configure_project` now store credentials in-memory first; disk write is best-effort and non-fatal (fixes failure when MCP server CWD is `/`)
- `ResolveClient` checks session before disk config; returns a helpful error guiding Claude to call `connect()` when nothing is configured
- `select_site` trims whitespace from `site_name` parameter
- `connect` correctly rejects empty or whitespace-only `token` and `api_url`
- `validMediaTypes` consolidated into a single definition in `internal/api` (was duplicated across `cli` and `mcp` packages)

## [0.3.0] - 2026-04-12

### Added

- `select_site` MCP tool — switch to a named site from global config mid-session

### Fixed

- `stqry completion zsh` no longer requires a configured site
- Renamed "Manage QR/NFC codes" to "Manage redemption codes"

### Documentation

- MCP server setup documented in README (Claude Code + Claude Desktop)

## [0.2.0] - 2026-04-12

### Added

- MCP server (`stqry mcp serve`) for AI agent integration
  - Full tool coverage: projects, collections, collection items, screens,
    sections, sub-items, media, and codes
  - MCP resources for reading projects, collections, screens, media, and codes
  - `configure_project` tool for setting active project context
  - `STQRY_SITE` environment variable support for site resolution
- `--version` / `-v` flags to display the CLI version

### Fixed

- S3 upload content-length handling
- Media upload enqueue flow
- Error responses now parsed as `[{code, message}]` (was `[]string`)
- Request bodies sent flat, not wrapped in an entity key
- Authorization headers in tests

## [0.1.0] - 2026-04-10

### Added

- CLI skeleton with global flags and help text
- Config management: `config init`, `config add-site`, `config list-sites`, `config switch`
  - Config stored in `stqry.yaml` / `stqry.yml`, resolved by walking up the directory tree
  - `--region` flag on `config add-site`
  - Inline credential support in `config init`
- API resources:
  - Projects (read-only)
  - Codes (full CRUD)
  - Collections and collection items (full CRUD)
  - Screens, sections, and sub-items — badges, links, media, prices, social, hours (full CRUD)
  - Media uploads via multi-step flow (presign → S3 upload → enqueue → poll)
- Output formatters: JSON, quiet/raw, and human-readable tables with translation display
- Homebrew installation support
- macOS code signing and notarization via `quill` (no Mac required in CI)
- Cross-platform binaries via GoReleaser: linux, darwin, windows × amd64/arm64
- Python SDK (`stqry` on PyPI) with full API coverage: Client, Projects, Collections,
  Screens, Media, and Codes resources
- GitHub Actions workflows for releases (Go binaries) and PyPI publishing (Python SDK)
