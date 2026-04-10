# Python SDK Design — `stqry` on PyPI

## Goal

Publish a `stqry` Python package to PyPI so that Claude chat (and other Python environments) can interact with the STQRY API via `pip install stqry` — no binary download required.

## Layout

A `python/` subdirectory in the existing repo, released on the same cadence as the Go CLI.

```
python/
  stqry/
    __init__.py       # re-exports Client
    _http.py          # raw HTTP client
    collections.py    # collections + collection_items
    screens.py        # screens + story_sections + section sub-items
    media.py          # media_items
    codes.py          # codes
    projects.py       # projects
  pyproject.toml
```

Adding a new endpoint = edit one file, add one method. No scaffolding changes.

## Client API

```python
import stqry

# Reads STQRY_API_URL + STQRY_API_TOKEN from env, or accept explicit args
client = stqry.Client()
client = stqry.Client(api_url="...", token="...")
```

Resources are plain attributes on the client:

```python
# Collections
client.collections.list(**query)
client.collections.get(id)
client.collections.create(**fields)
client.collections.update(id, **fields)
client.collections.delete(id)
client.collections.list_items(collection_id, **query)
client.collections.create_item(collection_id, **fields)
client.collections.update_item(collection_id, item_id, **fields)
client.collections.delete_item(collection_id, item_id)
client.collections.reorder_items(collection_id, item_ids)

# Screens
client.screens.list(**query)
client.screens.get(id)
client.screens.create(**fields)
client.screens.update(id, **fields)
client.screens.delete(id)
client.screens.list_sections(screen_id, **query)
client.screens.get_section(screen_id, section_id)
client.screens.create_section(screen_id, **fields)
client.screens.update_section(screen_id, section_id, **fields)
client.screens.delete_section(screen_id, section_id)
client.screens.reorder_sections(screen_id, section_ids)
client.screens.list_sub_items(screen_id, section_id, sub_item_type)
client.screens.create_sub_item(screen_id, section_id, sub_item_type, singular_key, **fields)
client.screens.update_sub_item(screen_id, section_id, sub_item_type, item_id, singular_key, **fields)
client.screens.delete_sub_item(screen_id, section_id, sub_item_type, item_id)

# Media
client.media.list(**query)
client.media.get(id)
client.media.create(**fields)
client.media.update(id, **fields)
client.media.delete(id, **query)

# Codes
client.codes.list(**query)
client.codes.get(id)
client.codes.create(**fields)
client.codes.update(id, **fields)
client.codes.delete(id)

# Projects
client.projects.list(**query)
client.projects.get(id)
```

All methods return `dict` or `list[dict]`. List methods accept `**query` kwargs as API query params. Mutation methods accept `**fields` as the request body.

## Dependencies

- `requests` — only dependency, always available in Claude's sandbox

## Authentication

Priority order:
1. Explicit `api_url` / `token` constructor args
2. `STQRY_API_URL` / `STQRY_API_TOKEN` environment variables

Raises `ValueError` at construction time if neither is provided.

## CI/CD

New `publish-pypi` job added to `.github/workflows/release.yml`, gated on the existing `goreleaser` job:

```yaml
publish-pypi:
  needs: [goreleaser]
  runs-on: ubuntu-latest
  environment: release
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0   # needed for vcs versioning
    - uses: actions/setup-python@v5
      with:
        python-version: "3.12"
    - run: pip install build twine
    - run: python -m build
      working-directory: python/
    - run: twine upload dist/*
      env:
        TWINE_USERNAME: __token__
        TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
```

Version is read dynamically from the git tag via `hatchling-vcs` — no manual version field in `pyproject.toml`.

One new secret required: `PYPI_TOKEN` in the `release` GitHub environment.
