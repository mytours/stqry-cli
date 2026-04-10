# stqry

Python SDK for the [STQRY](https://stqry.com) API.

## Installation

```bash
pip install stqry
```

## Usage

```python
import stqry

client = stqry.Client()  # reads STQRY_API_URL and STQRY_API_TOKEN from env
# or
client = stqry.Client(api_url="https://your-site.stqry.com", token="your-token")

# Collections
collections = client.collections.list()
collection = client.collections.get("123")

# Screens
screens = client.screens.list()

# Media
media = client.media.list()
```

## Authentication

Set environment variables:
- `STQRY_API_URL` — your STQRY site API URL
- `STQRY_API_TOKEN` — your API token
