# stqry

CLI for managing collections, screens, media, and content on STQRY.

## Installation

### Homebrew
```bash
brew install stqry/tap/stqry
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

STQRY CLI requires a site to be configured. Sites can be specified in two ways:

1. **Command flag** (takes precedence):
   ```bash
   stqry collection list --site mysite
   ```

2. **Configuration file** at `~/.stqry/config.yaml`:
   ```yaml
   site: mysite
   ```

If no site is configured, the command will prompt you to select one.
