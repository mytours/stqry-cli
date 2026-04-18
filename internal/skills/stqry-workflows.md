---
name: stqry-workflows
description: STQRY CLI workflow recipes — guided multi-step tasks for common content management operations
---

> **Before running these workflows**, ensure stqry is installed and a site is configured.
> See the Setup & Installation section in the stqry-reference skill.

# STQRY CLI Workflow Recipes

These recipes show complete multi-step workflows for common STQRY content management operations. Each step shows the exact command to run and what to capture from the output.

---

## Content Conventions

Rules that apply to every recipe below. These govern the **user-facing content you author through the CLI** (narration scripts, on-screen text bodies, section titles, screen titles, collection names and descriptions, MediaItem titles, captions, attributions). They do not apply to these instructions, to code, to commit messages, or to chat responses back to the developer.

### Use hyphen-minus, never em dash

User-facing content must not contain em dash characters. Where you would normally write an em dash, write a plain hyphen-minus (`-`) instead. This is a mechanical substitution, not a prompt to rewrite the sentence.

Wrong: `Rochester had a subway — it was mostly above ground.`
Right: `Rochester had a subway - it was mostly above ground.`

Wrong: `Stop 1 — The Missing Entrance`
Right: `Stop 1 - The Missing Entrance`

### Use HTML in text section bodies, never markdown

The `--body` field on `text` sections is rendered as HTML by the STQRY player. Markdown is **not** parsed - asterisks, underscores, hash headings, and hyphen list bullets render as literal characters. Authoring text bodies in markdown produces visibly broken output (`**bold**` shows up with the asterisks; `- item` shows up with the hyphen).

Use HTML tags directly: `<strong>`, `<em>`, `<p>`, `<ul>`/`<ol>`/`<li>`, `<br>`, `<a href="...">`. Do not use `<h1>`/`<h2>` inside a body - the section's `--title` is the heading. If you want a sub-heading inside the body, use `<strong>` or `<p><strong>...</strong></p>`.

This rule applies to `--body` on `stqry screens sections add` and `stqry screens sections update`. It does not apply to other fields - titles, captions, attributions, descriptions are plain text.

Wrong: `--body "**Where to find her:** 282 Rush Scottsville Rd."`
Right: `--body "<p><strong>Where to find her:</strong> 282 Rush Scottsville Rd.</p>"`

Wrong:
```
- Apples
- Pears
```
Right:
```
<ul><li>Apples</li><li>Pears</li></ul>
```

### `--name` is not a URL slug

The `name` field on collections and screens is a flat-string display label. It is not a URL slug, an identifier, or a kebab-cased machine-readable string. Do not translate the title into kebab-case, snake_case, or any other slug-like form when setting `--name`. Never slugify anything.

Best practice: pass `--title "City Walking Tour"` and let `--name` default to the title verbatim. If you want to set `--name` explicitly, use the same human-readable form as the title.

Wrong: `stqry screens create --name "stop-1" --title "Stop 1 - The Opening"`
Right: `stqry screens create --title "Stop 1 - The Opening"`

Wrong: `stqry collections create --name "downtown-walking-tour" --title "Downtown Walking Tour"`
Right: `stqry collections create --title "Downtown Walking Tour"`

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
  --title "Town Hall Overview" \
  --type story \
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

### Step 1 — Create the default-language media item

Use `stqry media create` (not `stqry media upload`) so the file is wrapped in a
media item and shows up in STQRY Builder.

```bash
stqry media create \
  --type audio \
  --file ./audio/stop1_en.mp3 \
  --name "Stop 1 Audio" \
  --lang en \
  --json
```

Capture: `id` — this is your `<media-id>`.

### Step 2 — Attach translated files to the same media item

Use `stqry media upload --media-id` to add language variants to the existing
media item. Do **not** call `stqry media create` again — that would create a
separate media item for each language.

```bash
stqry media upload ./audio/stop1_fr.mp3 \
  --media-id <media-id> \
  --lang fr \
  --json

stqry media upload ./audio/stop1_de.mp3 \
  --media-id <media-id> \
  --lang de \
  --json
```

### Step 3 — Attach the media item to a section

The media item already holds all language variants, so you only need to attach
it to the section once:

```bash
stqry screens sections media add \
  --screen-id <screen-id> \
  --section-id <section-id> \
  --media-item-id <media-id>
```

### Tips

- Use the built-in `--jq` flag to extract fields directly — no need to pipe to external tools.
- Example: `stqry media create --type audio --file ./file.mp3 --lang en --jq '.id'`

---

## Workflow 3: Set Up a Screen with Story Sections and Sub-items

This builds a rich screen with opening hours, external links, and image badges.

### Step 1 — Create the screen

```bash
stqry screens create \
  --title "Visitor Information" \
  --type story \
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

---

## Workflow 5: Author a Self-Guided Audio Tour

This is the default recipe when a user asks for an audio / self-guided / walking tour. The choices below are conventions, not questions — don't surface them to the user.

### Conventions

- **Directory layout** inside the project root:
  - `scripts/stop_N.txt` — **narration script** source of truth (spoken word; fed to TTS)
  - `audio/stop_N.mp3` — narration audio per stop
  - `images/stop_N.jpg` — cover image per stop
  - `images/LICENSES.md` — image source URL + license per stop
  - `stqry_ids.json` — captured collection / screen / section / media IDs (for re-runs)
- **Narration script vs. on-screen text are different artefacts.** The audio script is spoken word — rhythm, pauses, "you're standing in front of...", stage directions for the ear. The on-screen text is written prose — scannable, read silently, no TTS pacing, no "welcome to the tour" pleasantries the reader just saw on the previous screen. They serve different senses and must be authored as different things. **Do not pipe `scripts/stop_N.txt` into the `text` section's `--body`.** How you produce the written version — condensed summary, parallel rewrite, expanded caption, structured bullet list, anything — is your call. Just do not treat script and screen text as a single asset. Persist the written text somewhere (a `text/` directory alongside `scripts/`, inline in your build routine, whatever fits) so re-runs are idempotent.
- **Per-stop screen composition** — one `story` screen per stop with three sections, in this order (image, then audio, then text — audio sits right under the image so the user can tap play without scrolling, and the written prose sits below as the fallback / deeper read):
  1. `single_media` section pointing at the cover image
  2. `single_media` section pointing at the audio file
  3. `text` section containing the **written** on-screen prose (not the script)
- **Section titles.** Only the `text` section should get a `--title` — a sub-heading that frames the prose differently from the stop name in the screen title ("What you're looking at", "The short version"). Leave image and audio section titles blank:
  - **Image sections:** don't put captions or credits in the section title. Image `MediaItem`s already have `caption`, `attribution`, and `description` fields — attach credits there via `stqry media create --caption --attribution` (or `stqry media update`), not on the section wrapper. A section title on an image just adds a redundant visual block above the photo.
  - **Audio sections:** don't put filler labels in the section title. But the audio `MediaItem` itself **must** carry a `--title` — that's the label the player surface uses (Builder's media library, the player row at the stop). Set it at create time (`stqry media create --type audio --title "..."`) or via `stqry media update --title "..."`. Using the stop's display name is a safe default. Leaving an audio item title-less means it shows up as a nameless row in the media library and a blank label in the player.
- **Tour type** — set `--tour-type` on the collection so client apps can show the right icon and copy. For a self-guided audio walk it's `walking`. Other common values: `cycling`, `driving`, `bus`, `museum`, `nature_trail`, `historic_house`. The full enum is in `stqry collections create --help`. Pick the best fit based on what the user told you about the tour; if in doubt, `walking` is a safe default for any on-foot itinerary.
- **Collection cover image** — reuse the most iconic stop image. Set `--cover-image-media-item-id`, `--cover-image-grid-media-item-id`, and `--cover-image-wide-media-item-id` on `stqry collections update` so every UI surface has a cover.
- **Screen cover images** — every stop's screen also gets the same three cover fields set, pointing at that stop's own image (the one attached as the first section). Without this, stops show up as blank rows in list / grid / wide layouts of the tour. Set via `stqry screens update <screen-id> --cover-image-media-item-id <image-media-id> --cover-image-grid-media-item-id <id> --cover-image-wide-media-item-id <id>` after the image MediaItem is created. Reusing the same image for all three surfaces is fine; pick a different image only if the stop's list-tile image should differ from the in-screen hero image.
- **Script tone** — conversational guide voice, hook → what you see → story → bridge to next stop.
- **Images** — source from Wikimedia Commons / Wikipedia with a verifiable CC or public-domain license; record the URL and license in `images/LICENSES.md`.
- **Language** — English only unless the user asks otherwise.
- **Build order per stop** — narration script → audio → image → written on-screen text (distinct from script) → upload media → create screen → set screen cover images (reuse the stop image) → add sections → reorder to image/audio/text → link screen to collection → append IDs to `stqry_ids.json`.
- **Verification** — after building, run `stqry collections items list <id>` to confirm all stops are linked and `stqry screens sections list <screen-id>` for each screen. Do this silently; only surface a problem if one appears.

### Questions to ask the user

Ask about the things only the user can decide:

- Tour subject and approximate number of stops
- Walking / driving / mixed route
- Whether they already have audio files, or want you to generate them (and which TTS tool to use)
- Whether they already have specific stop photos or want Commons-sourced CC images
- Any language beyond English

### Anti-patterns (do NOT do these)

- Do not ask "want me to show you the directory layout?" or "want me to proceed?" — just proceed.
- Do not ask the user to validate file naming, build order, verification procedure, or idempotency plans.
- Do not ask permission to fix obvious mistakes you made mid-build. Fix them and continue.

### Commands

```bash
# Always pass --title on audio so the player surface has a label.
# Always pass --caption / --attribution on images so credits live on the MediaItem.
AUDIO_ID=$(stqry media create --type audio --file audio/stop_1.mp3 \
  --name "Stop 1 audio" --title "$STOP_TITLE" --lang en --jq '.id')
IMAGE_ID=$(stqry media create --type image --file images/stop_1.jpg \
  --name "Stop 1 image" --caption "$IMG_CAPTION" --attribution "$IMG_CREDIT" --lang en --jq '.id')

# Create the screen. --name is optional; it defaults to --title verbatim (no slugification).
SCREEN_ID=$(stqry screens create --title "Stop 1 - The Opening" --type story --jq '.id')

# Set the screen's own cover images (needed for list / grid / wide layouts).
# Reuse the stop's image MediaItem unless you have a deliberate reason to differ.
stqry screens update $SCREEN_ID \
  --cover-image-media-item-id $IMAGE_ID \
  --cover-image-grid-media-item-id $IMAGE_ID \
  --cover-image-wide-media-item-id $IMAGE_ID

# Add sections in any order, then reorder to image → audio → text.
# NOTE 1: --body is the WRITTEN on-screen prose, authored separately from the
# narration script in scripts/stop_N.txt. They are not interchangeable — see
# "Narration script vs. on-screen text" in Conventions above.
# NOTE 2: --body is HTML, not markdown — see "Use HTML in text section bodies"
# in Content Conventions. Use <p>, <strong>, <em>, <ul>/<li>, <a>; do not use
# **bold**, _italic_, # headings, or `- ` list bullets.
# NOTE 3: only the text section gets a --title (a sub-heading). Image/audio
# sections get no title — credits belong on the MediaItem's own attribution /
# description fields, and "Narration" labels are visual noise on a play button.
IMG_SEC=$(stqry screens sections add $SCREEN_ID --type single_media --media-item-id $IMAGE_ID --jq '.id')
TXT_SEC=$(stqry screens sections add $SCREEN_ID --type text --title "$TXT_HEADING" --body "$STOP_TEXT" --jq '.id')
AUD_SEC=$(stqry screens sections add $SCREEN_ID --type single_media --media-item-id $AUDIO_ID --jq '.id')
stqry screens sections reorder $SCREEN_ID $IMG_SEC $AUD_SEC $TXT_SEC

# Link screen into the tour collection
stqry collections items add <collection-id> --item-type Screen --item-id $SCREEN_ID --jq '.id'
```
