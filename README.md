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

Add a new site:

```bash
stqry add-site
```

Initialize configuration:

```bash
stqry config init
```

List collections:

```bash
stqry collection list
```

Create a screen:

```bash
stqry screen create
```

Upload media:

```bash
stqry media upload
```

## AI Agent Integration

Set up Claude AI agent integration for the current site:

```bash
stqry setup claude
```

Set up Claude AI agent integration globally:

```bash
stqry setup claude --global
```

## Output Formats

Commands support multiple output formats:

Default (human-readable):

```bash
stqry collection list
```

JSON output:

```bash
stqry collection list --json
```

Quiet mode (minimal output):

```bash
stqry collection list --quiet
```

Parse JSON output with jq:

```bash
stqry collection list --json | jq '.[] | .name'
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
