# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
