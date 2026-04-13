# AGENTS.md

This file helps AI agents use the STQRY CLI to read and manage STQRY content.
It covers the two integration paths, how to get connected, and where to find
detailed command references and workflow recipes.

## Setup in Claude Cowork

If `stqry` is not on PATH, install it via pip:

    pip install stqry

Verify with `stqry --version` before proceeding. Note: MCP server setup is not
available in Claude Cowork — use Path 1 (CLI subprocess) below instead.

## Integration Paths

### Path 1 — CLI subprocess

Run `stqry` commands directly as shell subprocesses. Works anywhere the binary
is installed. Best when you're executing a batch task, writing a script, or
working inside an environment where MCP isn't configured.

Always use `--quiet` for machine-readable output — it returns raw JSON with no
envelope, easy to pipe into `jq`:

    stqry collections list --quiet | jq '.[].id'

Use `--json` when you need pagination metadata alongside the data:

    stqry collections list --json | jq '.data[].id'

### Path 2 — MCP server

Run `stqry mcp serve` as a persistent MCP server over stdio. Best for
interactive agents that support the Model Context Protocol — gives you typed
tools, structured inputs, and session-level site selection without needing
to pass `--site` on every call.

Configure in `.claude/settings.json`:

```json
{
  "mcpServers": {
    "stqry": {
      "command": "stqry",
      "args": ["mcp", "serve"]
    }
  }
}
```

Or in Claude Desktop's `claude_desktop_config.json` with the same structure.

**MCP note:** Call the `connect` or `select_site` tool at the start of each
conversation to set the active site for the session — you won't need `--site`
on every subsequent call.

## Content Model

STQRY content is organised as a hierarchy:

```
Projects → Collections → Items → Screens → Sections
                                               └── Sub-items (hours, links, badges, media…)
Media Items  (standalone, reusable across screens)
Codes        (redemption codes linked to collections or screens)
```

When building content, the typical order is: create a collection, add items,
create screens on those items, add sections, attach media.

## Getting Connected

There is no default site. Every command requires a site to be configured via
one of these (highest priority first):

1. `--site <name>` flag
2. `stqry.yaml` in the current or any parent directory
3. Named entry in `~/.config/stqry/config.yaml`

Add a site to global config:

    stqry config add-site --name mysite --token <token> --region us

Pin a site to the current project directory:

    stqry config init --name mysite   # writes stqry.yaml

Available regions: `us`, `ca`, `eu`, `sg`, `au`

## Skills (Claude Code)

Install STQRY skill files into Claude Code for richer agent support:

    stqry setup claude          # current project (.claude/commands/)
    stqry setup claude --global # all projects (~/.claude/commands/)

Two skills are installed:

- **stqry-reference** — complete command reference, all flags, and data model
  relationships. Load it when you need to look up a specific command or flag.
- **stqry-workflows** — step-by-step workflow recipes for common multi-step
  tasks (create a tour, bulk upload media, manage translations). Load it when
  you're about to execute a multi-step content operation.

These skills are the authoritative reference. This file is the orientation —
go to the skills for specifics.

## Agent Guidelines

**Parsing output:** Use `--quiet` to get a bare JSON array or object — no
envelope, easy to extract IDs:

    stqry collections list --quiet | jq -r '.[0].id'

Use `--json` when you need pagination metadata (`meta.total`, `meta.page`).

**Chaining operations:** Capture IDs from each step before proceeding.
Collections, items, screens, and sections all return their `id` in the
response. Use `--quiet | jq -r '.id'` to extract cleanly.

**Errors:** The CLI exits non-zero on failure and writes a message to stderr.
API errors include a `code` and `message` field. Check exit code before using
output from a previous step.

**Pagination:** List commands default to page 1, 20 items. Use `--page` and
`--per-page` to walk results. With `--json`, check `meta.total` to know how
many pages exist.

**Verification:** Use `stqry doctor` to check connectivity and config before
starting a long content operation.
