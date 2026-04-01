---
name: stqry-reference
description: STQRY CLI command reference — all commands, flags, and data model relationships
---

# STQRY CLI Command Reference

## Data Model Overview

STQRY content is organised in a hierarchy:

```
Projects
  └── Collections
        └── Items
              └── Screens
                    └── Sections
                          └── Sub-items (hours, links, badges, prices, social, media)
Media Items  (standalone, attached to screens or sections)
Codes        (QR / NFC codes linked to collections or screens)
```

- **Projects** — top-level organisational units; each project belongs to one site.
- **Collections** — groupings of items (e.g. a tour, an exhibition). Each collection has a type (tour, exhibit, etc.).
- **Items** — individual entries within a collection (e.g. a stop on a tour).
- **Screens** — content pages attached to an item. An item can have multiple screens.
- **Sections** — content blocks within a screen (text, image, audio, video, etc.).
- **Sub-items** — structured data attached to a section: hours, links, badges, prices, social handles, media.
- **Media Items** — images, audio, video, and documents. Can be uploaded and attached independently.
- **Codes** — QR or NFC codes that deep-link into collections or screens.

---

## Site Configuration

There is **no global default site**. A site must be specified via:

1. `--site <name>` flag (highest priority)
2. `.stqry/config.yaml` in the current (or any parent) directory
3. A named site entry in `~/.config/stqry/config.yaml`

### Directory config (`.stqry/config.yaml`)

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

### Language Support

- When `--lang` is omitted the API returns the site's default language.
- Pass `--lang <code>` to retrieve or write content in a specific locale.
- Language codes follow IETF BCP 47 (e.g. `en`, `fr`, `de`, `zh-Hans`).

---

## Commands

### `stqry config`

Manage site configuration.

```
stqry config list                        List all configured sites
stqry config add <name>                  Add or update a site
stqry config remove <name>               Remove a site
stqry config show [<name>]               Show config for a site (or current)
```

---

### `stqry collections`

Manage collections and their items.

```
stqry collections list                   List collections
stqry collections get <id>               Get a single collection
stqry collections create                 Create a collection (interactive)
stqry collections update <id>            Update a collection
stqry collections delete <id>            Delete a collection

stqry collections items list <id>        List items in a collection
stqry collections items get <id>         Get a single item
stqry collections items create <col-id>  Create an item in a collection
stqry collections items update <id>      Update an item
stqry collections items delete <id>      Delete an item
```

---

### `stqry screens`

Manage screens and their sections / sub-items.

```
stqry screens list <item-id>             List screens for an item
stqry screens get <id>                   Get a single screen
stqry screens create <item-id>           Create a screen
stqry screens update <id>                Update a screen
stqry screens delete <id>                Delete a screen

# Sections
stqry screens sections list <screen-id>
stqry screens sections get <id>
stqry screens sections create <screen-id>
stqry screens sections update <id>
stqry screens sections delete <id>

# Sub-items (attached to a section)
stqry screens sections badges   list|get|create|update|delete
stqry screens sections links    list|get|create|update|delete
stqry screens sections media    list|get|create|update|delete
stqry screens sections prices   list|get|create|update|delete
stqry screens sections social   list|get|create|update|delete
stqry screens sections hours    list|get|create|update|delete
```

---

### `stqry media`

Manage and upload media items.

```
stqry media list                         List media items
stqry media get <id>                     Get a single media item
stqry media upload <file>                Upload a new media file
stqry media update <id>                  Update media metadata
stqry media delete <id>                  Delete a media item
```

Flags for `stqry media upload`:

| Flag | Description |
|------|-------------|
| `--title` | Title for the media item |
| `--lang` | Language of the media content |

---

### `stqry projects`

Manage projects.

```
stqry projects list                      List projects
stqry projects get <id>                  Get a single project
```

---

### `stqry codes`

Manage QR / NFC codes.

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
