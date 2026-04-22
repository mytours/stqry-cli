---
name: stqry-reference
description: STQRY CLI command reference — all commands, flags, and data model relationships
---

## Setup & Installation

### Claude Cowork

If `stqry` is not on PATH, install it via pip:

    pip install stqry

Verify with `stqry --version` before proceeding.

MCP server setup is not available in Claude Cowork — use CLI subprocess for all operations.

### First-Run Configuration (all environments)

Add a site to global config:

    stqry config add-site --name mysite --token <token> --region us

Available regions: `us`, `ca`, `eu`, `sg`, `au`

Write a project config file to the current directory:

    stqry config init --name mysite   # writes stqry.yaml

Verify connectivity:

    stqry doctor

**Claude Cowork persistence:** Download the generated `stqry.yaml` file and either
upload it at the start of future Cowork sessions, or add it to your Cowork working
folder so it is picked up automatically. This avoids reconfiguring each session.

---

# STQRY CLI Command Reference

## Data Model Overview

All content types are **top-level entities** — Screens, Collections, Media Items, Projects, and Codes exist independently and can be combined in different ways.

```
Top-level entities (all exist independently):
  Projects
  Collections
  Screens
    └── Sections
          └── Sub-items (hours, links, badges, prices, social, media)
  Media Items
  Codes

Collections are composed via Collection Items (join records):
  Collection
    └── Collection Items
          └── item_type: "Screen"  →  Screen
              item_type: "Collection"  →  Collection
```

**Collection Items are join records**, not content containers. A collection item has two fields: `item_type` (`"Screen"` or `"Collection"`) and `item_id` (the ID of the referenced entity). The referenced entity must exist before the collection item is created — create screens first, then link them into collections.

- **Projects** — top-level organisational units; each project belongs to one site.
- **Collections** — groupings of screens or other collections. Each collection has a type (tour, exhibit, etc.).
- **Screens** — standalone content pages. A screen has sections and can be linked into one or more collections.
- **Sections** — content blocks within a screen (text, image, audio, video, etc.).
- **Sub-items** — structured data attached to a section: hours, links, badges, prices, social handles, media.
- **Media Items** — images, audio, video, and documents. Can be uploaded and attached independently.
- **Codes** — redemption codes that deep-link into collections or screens.

---

## Site Configuration

There is **no global default site**. A site must be specified via:

1. `--site <name>` flag (highest priority)
2. `stqry.yaml` in the current (or any parent) directory
3. A named site entry in `~/.config/stqry/config.yaml`

### Directory config (`stqry.yaml`)

```yaml
site: my-site-name
```

### Global config (`~/.config/stqry/config.yaml`)

```yaml
sites:
  my-site-name:
    api_url: https://api.stqry.com
    token: <api-token>
  staging:
    api_url: https://api-staging.stqry.com
    token: <staging-token>
```

---

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--site` | string | Site name to use (overrides directory config) |
| `--lang` | string | Language code for content (e.g. `en`, `fr`, `de`) |
| `--json` | bool | Output full JSON response envelope |
| `--quiet` | bool | Output minimal JSON (no envelope) |
| `--jq` | string | Filter output with a jq expression (overrides `--quiet`) |
| `--progress` | bool | Show upload progress on stderr (off by default; `dd(1)`-style opt-in) |

### Extracting data with `--jq`

The CLI has **built-in jq filtering** via the `--jq` flag (powered by gojq). Always use `--jq` instead of piping to external `jq`, `python`, or other tools.

```bash
# List just screen names
stqry screens list --jq '.[].name'

# Get specific fields as objects
stqry collections list --jq '[.[] | {id, name, type}]'

# Filter by field value
stqry screens list --jq '[.[] | select(.type == "story")]'

# Count items
stqry media list --jq 'length'

# Get a single field from a get command
stqry screens get 12345 --jq '.name'
```

**Do NOT** pipe `--quiet` output into `python -c` or external `jq` — the built-in `--jq` flag is simpler and avoids extra dependencies.

### Bulk operations

There is no `set-all` or `apply-to-all` command. Combine `--jq` with a shell loop:

```bash
# Give every stop in a tour a 50m geofence.
for id in $(stqry collections items list 42 --jq '.[].id'); do
  stqry collections items update 42 "$id" --geofence gps --radius 50
done

# Nuke all geofences in a tour.
for id in $(stqry collections items list 42 --jq '.[].id'); do
  stqry collections items update 42 "$id" --geofence off
done
```

`update` with `--radius` merges into each item's existing `gps_settings`, so repeated calls don't clobber `geofence_lat` / `geofence_lng`.

### Map pin colour for a whole tour

Each tour stop's map pin colour lives on the collection item: `map_pin_colour` (CSS hex) alongside `map_pin_icon` and `map_pin_style`. Set via `--map-pin-colour` on `collections items update`. Validated client-side — the server requires a CSS hex (with or without leading `#`); free-text names like `red` return HTTP 422.

```bash
# Paint every stop in a tour with a brand colour.
for id in $(stqry collections items list 42 --jq '.[].id'); do
  stqry collections items update 42 "$id" --map-pin-colour "#FF6600"
done

# Reset every stop back to the tour default pin colour.
for id in $(stqry collections items list 42 --jq '.[].id'); do
  stqry collections items update 42 "$id" --map-pin-colour default
done
```

### Language Support

- When `--lang` is omitted the API returns the site's default language.
- Pass `--lang <code>` to retrieve or write content in a specific locale.
- Language codes follow IETF BCP 47 (e.g. `en`, `fr`, `de`, `zh-Hans`).

---

## Commands

### `stqry config`

Manage site configuration.

```
stqry config list-sites                                              List all configured sites
stqry config add-site --name <n> --token <t> --region <r>            Add a site to global config (region: us, ca, eu, sg, au; or pass --api-url for a custom endpoint)
stqry config edit-site <name> [--token <t>] [--api-url <url>]        Update an existing site's token or API URL
stqry config remove-site <name>                                      Remove a site from global config
stqry config init [--name <n>]                                       Write stqry.yaml in the current directory, pinning it to a site
stqry config show [<name>]                                           Show the fully resolved configuration with source tracking
stqry config validate                                                Check config file syntax and token validity
```

---

### `stqry collections`

Manage collections and their items.

```
stqry collections list                   List collections
stqry collections get <id>               Get a single collection
stqry collections create --type <type> [--name <n>] [--title <t>] [--short-title <t>] [--description <d>] [--tour-type <tt>]  Create a collection (--name or --title required). `name` and `title` are separate fields: `name` is a flat label, `title` is translatable. Pass --name for the label; add --title (with optional --lang) to populate the translatable title.
stqry collections update <id>            Update a collection
stqry collections delete <id>            Delete a collection

stqry collections items list <collection-id>                    List items in a collection
stqry collections items get <collection-id> <item-id>           Get a single collection item
stqry collections items add <collection-id> --item-type <type> --item-id <id> [--position <n>]  Add a screen or collection to a collection
stqry collections items update <collection-id> <item-id> [--position <n>] [--lat <l>] [--lng <l>] [--item-number <s>] [--geofence <mode>] [--radius <m>] [--geofence-lat <l>] [--geofence-lng <l>] [--geofence-content]  Update a single collection item (position, GPS pin, geofence). --radius merges into gps_settings so existing geofence_lat / geofence_lng aren't clobbered. Bulk with a shell loop — see the geofence snippet below.
stqry collections items reorder <collection-id> <item-id>...    Reorder items in a collection (1-based positions applied to the whole list)
stqry collections items remove <collection-id> <item-id>        Remove an item from a collection
```

---

### `stqry screens`

Manage screens and their sections / sub-items.

```
stqry screens list                                  List screens
stqry screens get <id>                              Get a single screen
stqry screens create --name <name> --type <type> [--title <t>] [--short-title <t>]    Create a screen. `name` (flat label) and `title` (translatable) are separate fields; pass --name for the label, and --title (with optional --lang) to populate the translatable title.
stqry screens update <id>                           Update a screen
stqry screens delete <id>                           Delete a screen

# Sections
stqry screens sections list <screen-id>
stqry screens sections get <section-id> --screen-id <screen-id>
stqry screens sections add <screen-id> --type <type>
stqry screens sections update <section-id> --screen-id <screen-id>
stqry screens sections remove <section-id> --screen-id <screen-id>
stqry screens sections reorder <screen-id> <section-id>...

# Sub-items (attached to a section; all require --screen-id and --section-id)
stqry screens sections badges   list|add|update|remove --screen-id <id> --section-id <id>
stqry screens sections links    list|add|update|remove --screen-id <id> --section-id <id>
stqry screens sections media    list|add|update|remove --screen-id <id> --section-id <id>
stqry screens sections prices   list|add|update|remove --screen-id <id> --section-id <id>
stqry screens sections social   list|add|update|remove --screen-id <id> --section-id <id>
stqry screens sections hours    list|add|update|remove --screen-id <id> --section-id <id>
```

---

### `stqry media`

Manage and upload media items.

```
stqry media list                                      List media items
stqry media get <id>                                  Get a single media item
stqry media create --type <type> --file <path>        Create a media item (uploads file if provided)
stqry media upload <file> --media-id <id> --lang <l>  Attach a new file to an EXISTING media item (--media-id required)
stqry media update <id>                               Update media metadata
stqry media delete <id>                               Delete a media item
```

**When the user asks to upload a file, use `stqry media create`** — not `stqry media upload`.

`stqry media upload` requires `--media-id` and only exists to attach a new file to a media item that already exists (e.g. replacing a file or adding a language variant). It will refuse to run without `--media-id`. Running without it is not supported because the resulting uploaded file would be orphaned — invisible in STQRY Builder and unlinkable from the CLI afterwards.

Flags for `stqry media create` and `stqry media update`:

| Flag | Description |
|------|-------------|
| `--type` | Media item type (required on create). One of: `map`, `webpackage`, `animation`, `audio`, `image`, `video`, `webvideo`, `ar`, `data` |
| `--file` | Path to the file to upload (create only) |
| `--name` | Media item name (internal label; not shown to end users) |
| `--caption` | **Image** caption (TranslatedString) |
| `--attribution` | **Image** credit line, e.g. "Photo by A. Borchert · CC BY-SA 4.0" (TranslatedString) |
| `--description` | **Image** long description (TranslatedString) |
| `--title` | **Audio** display title (TranslatedString) |
| `--transcription` | **Audio** transcription for accessibility (TranslatedString) |
| `--thumbnail-media-item-id` | **Audio / video** poster image — ID of an image MediaItem to render behind the audio / video in the player (pass `0` on update to clear) |
| `--lang` | Language code (global flag) — used as the locale key for all TranslatedString fields above |

**Put credits on the MediaItem, not on the section.** When attaching a CC/public-domain image to a tour stop, set `--caption` and `--attribution` on `stqry media create` / `update`. Do not put the credit in the enclosing section's `--title` — that wraps a redundant label around the image. Similarly, an audio section does not need a "Narration · 2 min" title; use `--title` on the audio MediaItem itself if a display title is needed.

Flags for `stqry media upload`:

| Flag | Description |
|------|-------------|
| `--media-id` | **Required.** Existing media item ID to attach the uploaded file to |
| `--lang` | **Required.** Language of the uploaded file (global flag) |

---

### `stqry projects`

Manage projects.

```
stqry projects list                      List projects
stqry projects get <id>                  Get a single project
```

---

### `stqry codes`

Manage redemption codes.

```
stqry codes list                         List codes
stqry codes get <id>                     Get a single code
stqry codes create                       Create a code
stqry codes update <id>                  Update a code
stqry codes delete <id>                  Delete a code
```

---

### `stqry setup`

Install tooling helpers.

```
stqry setup claude [--global]            Install Claude Code skill files
```

| Flag | Description |
|------|-------------|
| `--global` | Install to `~/.claude/commands/` instead of `./.claude/commands/` |
