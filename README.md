# stqry

CLI for managing collections, screens, media, and content on STQRY.

## Installation

### Homebrew

```bash
brew install mytours/tap/stqry-cli
```

### Go Install

```bash
go install github.com/mytours/stqry-cli/cmd/stqry@latest
```

### GitHub Releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/mytours/stqry-cli/releases).

## Quick Start

Add a site to your global config:

```bash
stqry config add-site --name=mysite --token=<API_TOKEN> --region=us
```

Pin that site to the current directory (writes `stqry.yaml`):

```bash
stqry config init --name=mysite
```

List collections:

```bash
stqry collections list
```

Create a screen (both `--name` and `--type` are required; `--type` must be
one of `story`, `web`, `panorama`, `ar`, `kiosk`):

```bash
stqry screens create --name="Welcome" --type=story
```

Create a media item from a local file (uploads the file and creates a media
item linked to it). `--type` must be one of `map`, `webpackage`, `animation`,
`audio`, `image`, `video`, `webvideo`, `ar`, `data`:

```bash
stqry media create --file=./photo.jpg --type=image --name=photo.jpg
```

## AI Agent Integration

### Skill Files (Claude Code)

Install STQRY workflow skills into Claude Code:

```bash
stqry setup claude          # current project
stqry setup claude --global # all projects
```

### MCP Server

`stqry mcp serve` starts an MCP server over stdio, letting AI assistants
read and manage STQRY content directly.

**Claude Code** — add to `.claude/settings.json` in your project:

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

The server picks up `stqry.yaml` from the project directory automatically.

**Claude Desktop** — add to `claude_desktop_config.json`:

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

No site is hardcoded. At the start of each conversation, tell Claude which
project to connect to:

- *"Use my site called mysite"* — if you've added it via `stqry config add-site`
- *"My token is `abc123`, region is US"* — Claude will configure it for you

## Output Formats

Commands support multiple output formats:

Default (human-readable):

```bash
stqry collections list
```

JSON output — wraps results in a `{ "data": [...], "meta": {...} }` envelope:

```bash
stqry collections list --json
```

Quiet mode — emits only the raw `data` payload (no envelope), handy for piping:

```bash
stqry collections list --quiet
```

Parse output with jq. Note that list endpoints are paginated — use `--page` /
`--per-page` to walk through results:

```bash
stqry collections list --quiet | jq '.[].name'
```

With `--json` you need to reach into the envelope:

```bash
stqry collections list --json | jq '.data[].name'
```

## Site Configuration

STQRY CLI requires a site to be configured. Sites are resolved in this order:

1. **Command flag** (highest priority):

   ```bash
   stqry collection list --site mysite
   ```

2. **Local folder config** — a `stqry.yaml` or `stqry.yml` file in the current directory (or any parent). Run `stqry config init` to create one:

   ```yaml
   site: mysite
   ```

   Or with inline credentials instead of a named site:

   ```yaml
   token: your-api-token
   api_url: https://api-us.stqry.com
   ```

   The CLI walks up the directory tree, so a `stqry.yaml` in a project root applies to all subdirectories.

3. **Global config** at `~/.config/stqry/config.yaml` — stores named sites added via `stqry config add-site`.

If no site is configured, the command will return an error prompting you to configure one.
