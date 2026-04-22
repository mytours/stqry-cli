# API Coverage

> 58 of 66 API operations covered (88%) — source: `docs/public_api.json`
>
> PATCH and PUT are collapsed into a single "update" operation throughout.
> ✅ = direct CLI command · ⚠️ = used internally · ❌ = not implemented

## Codes

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/codes` | `stqry codes list` | ✅ |
| POST   | `/api/public/codes` | `stqry codes create` | ✅ |
| GET    | `/api/public/codes/{id}` | `stqry codes get <id>` | ✅ |
| PATCH  | `/api/public/codes/{id}` | `stqry codes update <id>` | ✅ |
| DELETE | `/api/public/codes/{id}` | `stqry codes delete <id>` | ✅ |

## Collections

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/collections` | `stqry collections list` | ✅ |
| POST   | `/api/public/collections` | `stqry collections create` | ✅ |
| GET    | `/api/public/collections/{id}` | `stqry collections get <id>` | ✅ |
| PATCH  | `/api/public/collections/{id}` | `stqry collections update <id>` | ✅ |
| DELETE | `/api/public/collections/{id}` | `stqry collections delete <id>` | ✅ |

### Collection Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/collections/{id}/collection_items` | `stqry collections items list <id>` | ✅ |
| POST   | `/api/public/collections/{id}/collection_items` | `stqry collections items add <id>` | ✅ |
| POST   | `/api/public/collections/{id}/collection_items/update_positions` | `stqry collections items reorder <id>` | ✅ |
| GET    | `/api/public/collections/{id}/collection_items/{item_id}` | `stqry collections items get <id> <item_id>` | ✅ |
| PATCH  | `/api/public/collections/{id}/collection_items/{item_id}` | `stqry collections items update <id> <item_id>` | ✅ |
| DELETE | `/api/public/collections/{id}/collection_items/{item_id}` | `stqry collections items remove <id> <item_id>` | ✅ |

## Media Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/media_items` | `stqry media list` | ✅ |
| POST   | `/api/public/media_items` | `stqry media create` | ✅ |
| GET    | `/api/public/media_items/{id}` | `stqry media get <id>` | ✅ |
| PATCH  | `/api/public/media_items/{id}` | `stqry media update <id>` | ✅ |
| DELETE | `/api/public/media_items/{id}` | `stqry media delete <id>` | ✅ |

## Projects

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/projects` | `stqry projects list` | ✅ |
| GET    | `/api/public/projects/{id}` | `stqry projects get <id>` | ✅ |

## Screens

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens` | `stqry screens list` | ✅ |
| POST   | `/api/public/screens` | `stqry screens create` | ✅ |
| GET    | `/api/public/screens/{id}` | `stqry screens get <id>` | ✅ |
| PATCH  | `/api/public/screens/{id}` | `stqry screens update <id>` | ✅ |
| DELETE | `/api/public/screens/{id}` | `stqry screens delete <id>` | ✅ |

### Story Sections

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections` | `stqry screens sections list <screen-id>` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections` | `stqry screens sections add <screen-id>` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/update_positions` | `stqry screens sections reorder <screen-id>` | ✅ |
| GET    | `/api/public/screens/{id}/story_sections/{section_id}` | `stqry screens sections get <section-id>` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}` | `stqry screens sections update <section-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}` | `stqry screens sections remove <section-id>` | ✅ |

### Story Section Items

Sub-item commands take `--screen-id` and `--section-id` flags for list/add; update/remove take a positional `<item-id>`.

#### Badge Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/badge_items` | `stqry screens sections badges list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/badge_items` | `stqry screens sections badges add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/badge_items/{item_id}` | `stqry screens sections badges update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/badge_items/{item_id}` | `stqry screens sections badges remove <item-id>` | ✅ |

#### Link Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/link_items` | `stqry screens sections links list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/link_items` | `stqry screens sections links add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/link_items/{item_id}` | `stqry screens sections links update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/link_items/{item_id}` | `stqry screens sections links remove <item-id>` | ✅ |

#### Media Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/media_items` | `stqry screens sections media list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/media_items` | `stqry screens sections media add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/media_items/{item_id}` | `stqry screens sections media update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/media_items/{item_id}` | `stqry screens sections media remove <item-id>` | ✅ |

#### Opening Time Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/opening_time_items` | `stqry screens sections hours list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/opening_time_items` | `stqry screens sections hours add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/opening_time_items/{item_id}` | `stqry screens sections hours update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/opening_time_items/{item_id}` | `stqry screens sections hours remove <item-id>` | ✅ |

#### Price Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/price_items` | `stqry screens sections prices list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/price_items` | `stqry screens sections prices add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/price_items/{item_id}` | `stqry screens sections prices update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/price_items/{item_id}` | `stqry screens sections prices remove <item-id>` | ✅ |

#### Social Items

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| GET    | `/api/public/screens/{id}/story_sections/{section_id}/social_items` | `stqry screens sections social list` | ✅ |
| POST   | `/api/public/screens/{id}/story_sections/{section_id}/social_items` | `stqry screens sections social add` | ✅ |
| PATCH  | `/api/public/screens/{id}/story_sections/{section_id}/social_items/{item_id}` | `stqry screens sections social update <item-id>` | ✅ |
| DELETE | `/api/public/screens/{id}/story_sections/{section_id}/social_items/{item_id}` | `stqry screens sections social remove <item-id>` | ✅ |

## Uploaded Files

> These endpoints support the `stqry media upload` workflow. The presigned/process endpoints are called internally; direct CRUD commands are not yet implemented.

| Method | Endpoint | CLI Command | Status |
|--------|----------|-------------|--------|
| POST   | `/api/public/uploaded_files/presigned` | — | ⚠️ internal — `stqry media upload` |
| POST   | `/api/public/uploaded_files/process_enqueue` | — | ⚠️ internal — `stqry media upload` |
| GET    | `/api/public/uploaded_files/process_status/{job_id}` | — | ⚠️ internal — `stqry media upload` |
| GET    | `/api/public/uploaded_files` | — | ❌ |
| POST   | `/api/public/uploaded_files` | — | ❌ |
| GET    | `/api/public/uploaded_files/{id}` | — | ❌ |
| PATCH  | `/api/public/uploaded_files/{id}` | — | ❌ |
| DELETE | `/api/public/uploaded_files/{id}` | — | ❌ |

## Future Work

Unimplemented endpoints, in priority order:

- **`stqry files list`** — GET `/uploaded_files`
- **`stqry files get <id>`** — GET `/uploaded_files/{id}`
- **`stqry files update <id>`** — PATCH `/uploaded_files/{id}`
- **`stqry files delete <id>`** — DELETE `/uploaded_files/{id}`
- **`stqry files create`** — POST `/uploaded_files` (record only; uploading content uses `stqry media upload`)

## Keeping This Up To Date

When adding a CLI command that maps to a public API endpoint:

1. Find the endpoint row in the table above
2. Add the CLI command and change status to ✅
3. Update the coverage count in the header
4. Remove the entry from Future Work if present
