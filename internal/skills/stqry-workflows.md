---
name: stqry-workflows
description: STQRY CLI workflow recipes — guided multi-step tasks for common content management operations
---

# STQRY CLI Workflow Recipes

These recipes show complete multi-step workflows for common STQRY content management operations. Each step shows the exact command to run and what to capture from the output.

---

## Workflow 1: Create a New Tour (Collection + Screens + Items + Sections)

A "tour" is a Collection that links to Screens via collection items. Screens are standalone entities — create them first, then link each one into the collection. A collection item is a join record with two fields: `item_type` (e.g. "Screen") and `item_id` (the ID of the screen being linked).

### Step 1 — Create the collection

```bash
stqry collections create \
  --name "City Walking Tour" \
  --type tour \
  --json
```

Capture: `id` from the response — this is your `<collection-id>`.

### Step 2 — Create each screen (one per stop)

Screens are created as standalone entities before they are linked into any collection.

```bash
stqry screens create \
  --name "town-hall-overview" \
  --type story \
  --title "Town Hall Overview" \
  --json
```

Capture: `id` from each screen — this is your `<screen-id>`.

Repeat for every stop in the tour.

### Step 3 — Link each screen into the collection

A collection item is purely a link record — it has no content of its own. Use `collections items add` to attach each screen to the collection.

```bash
stqry collections items add <collection-id> \
  --item-type Screen \
  --item-id <screen-id> \
  --json
```

Repeat for each screen. Note that `collections items add` does not control position — use `stqry collections items reorder` afterwards to set the order screens appear in the tour.

### Step 4 — Add sections to each screen

```bash
# Text section
stqry screens sections add <screen-id> \
  --type text \
  --title "Town Hall Overview" \
  --json

# Image section — create the section, then attach media to it
stqry screens sections add <screen-id> \
  --type image \
  --json
# Capture: <section-id> from the response above

stqry screens sections media add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --media-item-id <media-id>
```

Repeat for additional content blocks.

---

## Workflow 2: Bulk Upload Media with Translations

Upload the same asset in multiple languages, then attach each to the appropriate section.

### Step 1 — Upload the default-language file

```bash
stqry media upload ./audio/stop1_en.mp3 \
  --title "Stop 1 Audio" \
  --lang en \
  --json
```

Capture: `id` — this is your `<media-id-en>`.

### Step 2 — Upload translated files

```bash
stqry media upload ./audio/stop1_fr.mp3 \
  --title "Stop 1 Audio" \
  --lang fr \
  --json

stqry media upload ./audio/stop1_de.mp3 \
  --title "Stop 1 Audio" \
  --lang de \
  --json
```

Capture each `id`.

### Step 3 — Attach to a section by language

```bash
stqry screens sections media add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --media-item-id <media-id-en> \
  --lang en

stqry screens sections media add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --media-item-id <media-id-fr> \
  --lang fr

stqry screens sections media add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --media-item-id <media-id-de> \
  --lang de
```

### Tips

- Use the built-in `--jq` flag to extract fields directly — no need to pipe to external tools.
- Example: `stqry media upload ./file.mp3 --jq '.id'`

---

## Workflow 3: Set Up a Screen with Story Sections and Sub-items

This builds a rich screen with opening hours, external links, and image badges.

### Step 1 — Create the screen

```bash
stqry screens create \
  --name "visitor-information" \
  --type story \
  --title "Visitor Information" \
  --json
```

Capture: `<screen-id>`.

### Step 2 — Add a text intro section

```bash
stqry screens sections add <screen-id> \
  --type text \
  --title "Plan Your Visit" \
  --json
```

Capture: `<section-id>`.

### Step 3 — Add opening hours to the section

```bash
stqry screens sections hours add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --description "Monday" \
  --time "09:00-17:00"

stqry screens sections hours add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --description "Tuesday" \
  --time "09:00-17:00"

# Repeat for each day; omit closed days
```

### Step 4 — Add external links

```bash
stqry screens sections links add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --link-type web \
  --url "https://tickets.example.com" \
  --label "Book Tickets"

stqry screens sections links add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --link-type web \
  --url "https://example.com" \
  --label "Official Website"
```

### Step 5 — Add a badge

```bash
stqry screens sections badges add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --badge-id <badge-id>
```

---

## Workflow 4: Manage Content Across Languages

Update content for a screen or section in multiple locales.

### Step 1 — View current content in the default language

```bash
stqry screens get <screen-id> --json
```

### Step 2 — Update the screen title in French

```bash
stqry screens update <screen-id> \
  --lang fr \
  --title "Informations pour les visiteurs"
```

### Step 3 — Update section title per language

```bash
stqry screens sections update <section-id> \
  --screen-id <screen-id> \
  --lang fr \
  --title "Informations pratiques"

stqry screens sections update <section-id> \
  --screen-id <screen-id> \
  --lang de \
  --title "Besucherinformationen"
```

### Step 4 — Verify translations

```bash
stqry screens get <screen-id> --lang fr --json
stqry screens get <screen-id> --lang de --json
```

### Tips for multilingual content management

- Always update the default language first to establish the canonical content.
- Use `--jq` to extract or reshape output inline (e.g. `--jq '.title'`), or `--quiet` for compact JSON suitable for diffing or logging.
- When scripting bulk translations, iterate over language codes and wrap each `update` in an error check.
- Language codes follow IETF BCP 47: `en`, `fr`, `de`, `es`, `zh-Hans`, `zh-Hant`, `pt-BR`, etc.
