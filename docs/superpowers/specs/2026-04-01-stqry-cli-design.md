# STQRY CLI Design Spec

**Story:** [#139678 — Add stqry-public-cli command line tool](https://app.shortcut.com/oncell/story/139678)
**Date:** 2026-04-01

## Overview

A cross-platform command-line tool (`stqry`) that allows STQRY customers and staff to manage collections, screens, media, and content from the terminal. Built as a Go binary for zero-dependency installation on Windows, macOS, and Linux.

**Primary audience:** Content managers and staff managing tours, screens, and media.
**Secondary audience:** Developers automating content pipelines and integrations.

## Technology

- **Language:** Go
- **CLI framework:** Cobra (command structure) + Viper (configuration)
- **Distribution:** GoReleaser for cross-platform binaries, Homebrew (`stqry/tap/stqry`), GitHub Releases, `go install`
- **API:** STQRY Public API (`/api/public/`)

## Authentication & Configuration

### Site Configuration

Users configure named sites with API tokens. There is no global default site — every command must resolve a site explicitly to prevent accidental operations against the wrong site.

**Global config file:** `~/.config/stqry/config.yaml`

```yaml
sites:
  bobs:
    token: ABC123
    api_url: https://api-us.area360.com
  museum-fr:
    token: DEF456
    api_url: https://api-ca.area360.com
```

**Directory config file:** `.stqry/config.yaml` (in project directory)

```yaml
site: bobs
```

### Site Resolution Order

1. `--site` flag (explicit override)
2. `.stqry/config.yaml` found in current directory or walking up parent directories
3. Error: "No site specified. Use `--site` or run `stqry config init` in this directory."

### Config Commands

```
stqry config add-site --name=<name> --token=<token> --api-url=<url>
stqry config remove-site <name>
stqry config edit-site <name> [--token=<token>] [--api-url=<url>]
stqry config list-sites
stqry config init    # creates .stqry/config.yaml in current directory, prompts for site name
```

## Command Structure

### Global Flags

| Flag | Description |
|------|-------------|
| `--site=<name>` | Override active site |
| `--lang=<code>` | Language context for write operations (defaults to resource's primary_language) |
| `--json` | Structured JSON output with metadata envelope |
| `--quiet` | Raw JSON data only, no envelope |

### Collections

Collections contain collection items, which link to screens or other collections.

```
stqry collections list
stqry collections get <id>
stqry collections create --name=<name> --type=list|tour|organization|menu|search [--title=<title>] ...
stqry collections update <id> [--name=<name>] [--title=<title>] ...
stqry collections delete <id>

stqry collections items list <collection-id>
stqry collections items add <collection-id> --item-type=screen|collection --item-id=<id>
stqry collections items update <collection-id> <item-id> ...
stqry collections items remove <collection-id> <item-id>
stqry collections items reorder <collection-id> <item-id>... # ordered list of IDs
```

### Screens

Screens contain story sections. Screen types: story, web, panorama, ar, kiosk.

```
stqry screens list
stqry screens get <id>
stqry screens create --name=<name> --type=story|web|panorama|ar|kiosk ...
stqry screens update <id> [--name=<name>] [--title=<title>] ...
stqry screens delete <id>
```

### Screen Sections

Story sections belong to a screen. Section IDs are globally unique — commands that operate on a specific section only need the section-id. Commands that list or add sections require the parent screen-id.

Section types: text, single_media, media_group, link_group, social_group, image_slider, location, menu, quiz_question, quiz_score, form.

```
stqry screens sections list <screen-id>
stqry screens sections get <section-id>
stqry screens sections add <screen-id> --type=<type> ...
stqry screens sections update <section-id> ...
stqry screens sections remove <section-id>
stqry screens sections reorder <screen-id> <section-id>... # ordered list of IDs
```

### Section Sub-Items

Each sub-item type is a subcommand under `stqry screens sections`. These are tied to specific section types (e.g., opening hours only on location sections, prices only on menu sections). The CLI validates that the section type supports the sub-item.

**Badges:**
```
stqry screens sections badges list <section-id>
stqry screens sections badges add <section-id> --name=<name> ...
stqry screens sections badges update <badge-id> ...
stqry screens sections badges remove <badge-id>
```

**Links:**
```
stqry screens sections links list <section-id>
stqry screens sections links add <section-id> --url=<url> ...
stqry screens sections links update <link-id> ...
stqry screens sections links remove <link-id>
```

**Media (section-level):**
```
stqry screens sections media list <section-id>
stqry screens sections media add <section-id> --media-id=<id>
stqry screens sections media update <media-item-id> ...
stqry screens sections media remove <media-item-id>
```

**Prices:**
```
stqry screens sections prices list <section-id>
stqry screens sections prices add <section-id> --name=<name> --price=<amount> ...
stqry screens sections prices update <price-id> ...
stqry screens sections prices remove <price-id>
```

**Social:**
```
stqry screens sections social list <section-id>
stqry screens sections social add <section-id> --platform=<platform> --url=<url> ...
stqry screens sections social update <social-id> ...
stqry screens sections social remove <social-id>
```

**Opening Hours:**
```
stqry screens sections hours list <section-id>
stqry screens sections hours add <section-id> --day=<day> --open=<time> --close=<time> ...
stqry screens sections hours update <hours-id> ...
stqry screens sections hours remove <hours-id>
```

### Media

Media items are standalone, reusable resources that can be attached to screens, collections, sections, etc. They support multiple file types (image, audio, video, ar, webpackage, map, animation, webvideo, data) and multi-language file translations.

```
stqry media list [--q=<search>]
stqry media get <id>
stqry media create --file=<path> --type=image|audio|video|... [--name=<name>] [--lang=<code>]
stqry media update <id> [--name=<name>] ...
stqry media delete <id> [--lang=<code>]
stqry media upload <file> [--lang=<code>] [--media-id=<id>]
```

### Projects

Read-only access to project information.

```
stqry projects list
stqry projects get <id>
```

### Codes

Codes are coupon/access codes linked to a Project or Collection.

```
stqry codes list
stqry codes get <id>
stqry codes create --coupon-code=<code> --linked-type=project|collection --linked-id=<id> [--valid-from=<date>] [--valid-to=<date>] [--max-redemptions=<n>]
stqry codes update <id> [--coupon-code=<code>] [--valid-from=<date>] [--valid-to=<date>] [--max-redemptions=<n>]
stqry codes delete <id>
```

### Setup

Install AI agent integrations.

```
stqry setup claude [--global]    # install Claude Code skill
```

- Without `--global`: installs to `.claude/` in the current directory (project-level)
- With `--global`: installs to `~/.claude/` (available in all conversations)

The installed skill includes:
1. **Command reference** — all available commands, flags, data model relationships (collections > items, screens > sections > sub-items, standalone media)
2. **Workflow skills** — guided multi-step recipes for common tasks:
   - Create a new tour (collection + screens + sections + media)
   - Bulk upload media with translations
   - Manage content across languages
   - Set up a screen with story sections and sub-items

The skill files are embedded in the Go binary and written to disk during setup.

## Media Upload Flow

Media upload is a multi-step orchestrated process with progress feedback.

### Upload-only (`stqry media upload <file>`)

1. `POST /api/public/uploaded_files/presigned` — get presigned S3 URL and upload parameters
2. Upload file to presigned URL — display progress bar with percentage and speed
3. `POST /api/public/uploaded_files/process_enqueue` — start server-side processing
4. Poll `GET /api/public/uploaded_files/process_status/{job_id}` — wait for processing to complete
5. Return the uploaded_file record

### Create media item (`stqry media create --file=<path> --type=image`)

1. Run the upload flow above
2. `POST /api/public/media_items` with `file_uploaded_file_id` set to the uploaded file ID
3. Return the media_item record

### Add translation file (`stqry media upload <file> --lang=fr --media-id=456`)

1. Run the upload flow
2. `PATCH /api/public/media_items/456` with `file_uploaded_file_id: {"fr": <new_uploaded_file_id>}`
3. Return updated media_item record

### Upload Flags

| Flag | Description |
|------|-------------|
| `--lang=<code>` | Language for the file (defaults to resource's primary_language) |
| `--media-id=<id>` | Attach uploaded file to existing media item as a translation |
| `--timeout=<duration>` | Max time to wait for processing (default 5m) |

## Language / Translation Support

STQRY resources support multi-language content. Translatable fields use locale maps (e.g., `{"en": "English text", "fr": "Texte francais"}`). Media items support per-language file uploads.

- Every resource has `primary_language` and `translated_languages` fields
- The `--lang` flag on write operations specifies which language to write to
- When `--lang` is omitted, writes target the resource's `primary_language`
- List/get commands return all translations by default
- `stqry media delete <id> --lang=fr` removes only the French translation

## Output Modes

### Human (default)

Formatted tables for list commands, key-value display for get commands. Translated fields shown as `[en] Title here [fr] Titre ici`. Colors for status indicators and headers.

### JSON (`--json`)

Full response envelope with metadata:

```json
{
  "data": { ... },
  "meta": {
    "site": "bobs",
    "page": 1,
    "per_page": 25,
    "total": 42
  }
}
```

### Quiet (`--quiet`)

Raw JSON data only — the `data` field contents without the envelope. Suitable for `jq` pipelines and scripting.

### Pagination

- List commands default to first page
- `--page=<n>` and `--per-page=<n>` flags for manual pagination
- `--all` flag to auto-paginate and return all results

### Errors

- Human-readable error messages to stderr in default mode
- Structured error JSON in `--json` mode
- Clear messages for common issues: expired/invalid token, site not configured, missing permissions, resource not found

## Project Structure

```
stqry-cli/
├── cmd/
│   └── stqry/
│       └── main.go                  # entry point
├── internal/
│   ├── cli/                         # Cobra command definitions
│   │   ├── root.go                  # root command, global flags
│   │   ├── config.go                # config subcommands
│   │   ├── collections.go           # collections + items subcommands
│   │   ├── screens.go               # screens + sections + sub-items
│   │   ├── media.go                 # media + upload
│   │   ├── projects.go              # projects
│   │   └── codes.go                 # codes
│   ├── api/                         # HTTP client and API methods
│   │   ├── client.go                # base client, auth headers, error handling
│   │   ├── collections.go           # collection & collection_item endpoints
│   │   ├── screens.go               # screen, section, sub-item endpoints
│   │   ├── media.go                 # media_item endpoints
│   │   ├── upload.go                # presign, upload, enqueue, poll flow
│   │   ├── projects.go              # project endpoints
│   │   └── codes.go                 # code endpoints
│   ├── config/                      # config loading and site resolution
│   │   └── config.go
│   └── output/                      # table, JSON, quiet formatters
│       └── output.go
├── .goreleaser.yaml                 # cross-platform build config
├── go.mod
├── go.sum
├── Makefile                         # build, test, lint targets
└── README.md
```

## Distribution

- **GitHub Releases:** Cross-platform binaries via GoReleaser (darwin/linux/windows, amd64/arm64)
- **Homebrew:** `brew install stqry/tap/stqry`
- **Direct download:** curl one-liner documented in README
- **Go install:** `go install github.com/mytours/stqry-cli/cmd/stqry@latest`

## Testing

- **Unit tests:** Config resolution logic, output formatting, API client request building
- **Integration tests:** Mock HTTP server for full API flows including multi-step upload
- **Build targets:** `make build`, `make test`, `make lint`
