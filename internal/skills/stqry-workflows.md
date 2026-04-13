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
stqry screens sections create <screen-id> \
  --type text \
  --body "Built in 1892, the Town Hall is..." \
  --json

# Image section
stqry screens sections create <screen-id> \
  --type image \
  --media-id <media-id> \
  --json
```

Repeat for additional content blocks. Sections are ordered by their `position` field.

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
stqry screens sections media create <section-id> \
  --media-id <media-id-en> \
  --lang en

stqry screens sections media create <section-id> \
  --media-id <media-id-fr> \
  --lang fr

stqry screens sections media create <section-id> \
  --media-id <media-id-de> \
  --lang de
```

### Tips

- Use `--quiet` to get minimal JSON output when scripting, making it easier to pipe to `jq`.
- Example: `stqry media upload ./file.mp3 --quiet | jq -r '.id'`

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
stqry screens sections create <screen-id> \
  --type text \
  --heading "Plan Your Visit" \
  --body "Open daily. Guided tours run at 10am and 2pm." \
  --position 1 \
  --json
```

Capture: `<section-id>`.

### Step 3 — Add opening hours to the section

```bash
stqry screens sections hours create <section-id> \
  --day monday \
  --open "09:00" \
  --close "17:00"

stqry screens sections hours create <section-id> \
  --day tuesday \
  --open "09:00" \
  --close "17:00"

# Repeat for each day; omit closed days or set --closed flag
```

### Step 4 — Add external links

```bash
stqry screens sections links create <section-id> \
  --label "Book Tickets" \
  --url "https://tickets.example.com" \
  --position 1

stqry screens sections links create <section-id> \
  --label "Official Website" \
  --url "https://example.com" \
  --position 2
```

### Step 5 — Add a badge (icon + label overlay)

```bash
stqry screens sections badges create <section-id> \
  --label "Free Entry" \
  --media-id <badge-icon-media-id>
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

### Step 3 — Update section body text per language

```bash
stqry screens sections update <section-id> \
  --lang fr \
  --body "Ouvert tous les jours. Visites guidées à 10h et 14h."

stqry screens sections update <section-id> \
  --lang de \
  --body "Täglich geöffnet. Führungen um 10 und 14 Uhr."
```

### Step 4 — Verify translations

```bash
stqry screens get <screen-id> --lang fr --json
stqry screens get <screen-id> --lang de --json
```

### Tips for multilingual content management

- Always update the default language first to establish the canonical content.
- Use `--quiet --json` together to get compact output suitable for diffing or logging.
- When scripting bulk translations, iterate over language codes and wrap each `update` in an error check.
- Language codes follow IETF BCP 47: `en`, `fr`, `de`, `es`, `zh-Hans`, `zh-Hant`, `pt-BR`, etc.
