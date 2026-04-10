# Python SDK Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a `stqry` Python package in `python/` that wraps the STQRY REST API and can be installed via `pip install stqry`.

**Architecture:** A thin HTTP layer (`_http.py`) is injected into resource classes (`collections.py`, `screens.py`, etc.). The `Client` class in `__init__.py` wires them together and resolves credentials from env vars or explicit args.

**Tech Stack:** Python 3.9+, `requests`, `hatchling` + `hatch-vcs` for packaging, `pytest` for tests.

---

### Task 1: Package scaffold

**Files:**
- Create: `python/pyproject.toml`
- Create: `python/stqry/__init__.py` (empty stub)

**Step 1: Create `python/pyproject.toml`**

```toml
[build-system]
requires = ["hatchling", "hatch-vcs"]
build-backend = "hatchling.build"

[project]
name = "stqry"
dynamic = ["version"]
description = "Python SDK for the STQRY API"
requires-python = ">=3.9"
dependencies = ["requests>=2.28"]

[project.optional-dependencies]
dev = ["pytest>=7"]

[tool.hatch.version]
source = "vcs"
raw-options = { root = ".." }

[tool.hatch.build.targets.wheel]
packages = ["stqry"]
```

**Step 2: Create `python/stqry/__init__.py`** (empty for now)

```python
```

**Step 3: Verify the package installs**

```bash
cd python && pip install -e ".[dev]"
```

Expected: installs without error.

**Step 4: Commit**

```bash
git add python/
git commit -m "feat(python): scaffold stqry package"
```

---

### Task 2: HTTP client

**Files:**
- Create: `python/stqry/_http.py`
- Create: `python/tests/__init__.py` (empty)
- Create: `python/tests/test_http.py`

**Step 1: Write the failing test**

```python
# python/tests/test_http.py
from unittest.mock import MagicMock, patch
from stqry._http import HttpClient


def _make_client():
    return HttpClient("https://example.com", "tok123")


def test_get_calls_correct_url():
    client = _make_client()
    mock_resp = MagicMock()
    mock_resp.content = b'{"foo": "bar"}'
    mock_resp.json.return_value = {"foo": "bar"}
    client._session.get = MagicMock(return_value=mock_resp)

    result = client.get("/api/public/collections")

    client._session.get.assert_called_once_with(
        "https://example.com/api/public/collections", params=None, timeout=30
    )
    assert result == {"foo": "bar"}


def test_post_wraps_json():
    client = _make_client()
    mock_resp = MagicMock()
    mock_resp.content = b'{"id": "1"}'
    mock_resp.json.return_value = {"id": "1"}
    client._session.post = MagicMock(return_value=mock_resp)

    result = client.post("/api/public/collections", {"collection": {"name": "Foo"}})

    client._session.post.assert_called_once_with(
        "https://example.com/api/public/collections",
        json={"collection": {"name": "Foo"}},
        timeout=30,
    )
    assert result == {"id": "1"}


def test_delete_returns_none():
    client = _make_client()
    mock_resp = MagicMock()
    mock_resp.content = b""
    client._session.delete = MagicMock(return_value=mock_resp)

    result = client.delete("/api/public/collections/1")

    assert result is None


def test_auth_header_is_set():
    client = _make_client()
    assert client._session.headers["X-Api-Token"] == "tok123"
```

**Step 2: Run to verify it fails**

```bash
cd python && pytest tests/test_http.py -v
```

Expected: `ImportError: cannot import name 'HttpClient'`

**Step 3: Implement `python/stqry/_http.py`**

```python
import requests


class HttpClient:
    def __init__(self, api_url: str, token: str):
        self._base_url = api_url.rstrip("/")
        self._session = requests.Session()
        self._session.headers.update({
            "X-Api-Token": token,
            "Content-Type": "application/json",
            "Accept": "application/json",
        })

    def get(self, path: str, params: dict = None) -> dict:
        resp = self._session.get(self._base_url + path, params=params, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def post(self, path: str, json: dict = None) -> dict:
        resp = self._session.post(self._base_url + path, json=json, timeout=30)
        resp.raise_for_status()
        return resp.json() if resp.content else {}

    def patch(self, path: str, json: dict) -> dict:
        resp = self._session.patch(self._base_url + path, json=json, timeout=30)
        resp.raise_for_status()
        return resp.json()

    def delete(self, path: str, params: dict = None) -> None:
        resp = self._session.delete(self._base_url + path, params=params, timeout=30)
        resp.raise_for_status()
```

**Step 4: Run tests**

```bash
cd python && pytest tests/test_http.py -v
```

Expected: all 4 tests pass.

**Step 5: Commit**

```bash
git add python/stqry/_http.py python/tests/
git commit -m "feat(python): add HTTP client"
```

---

### Task 3: Client class

**Files:**
- Modify: `python/stqry/__init__.py`
- Create: `python/tests/test_client.py`

**Step 1: Write the failing test**

```python
# python/tests/test_client.py
import os
import pytest
from unittest.mock import patch
from stqry import Client


def test_client_reads_env_vars():
    with patch.dict(os.environ, {"STQRY_API_URL": "https://api.example.com", "STQRY_API_TOKEN": "tok"}):
        client = Client()
    assert client.collections is not None
    assert client.screens is not None
    assert client.media is not None
    assert client.codes is not None
    assert client.projects is not None


def test_client_explicit_args_override_env():
    with patch.dict(os.environ, {"STQRY_API_URL": "https://wrong.com", "STQRY_API_TOKEN": "wrong"}):
        client = Client(api_url="https://right.com", token="correct")
    assert client._http._base_url == "https://right.com"
    assert client._http._session.headers["X-Api-Token"] == "correct"


def test_client_raises_if_no_credentials():
    env = {k: v for k, v in os.environ.items() if k not in ("STQRY_API_URL", "STQRY_API_TOKEN")}
    with patch.dict(os.environ, env, clear=True):
        with pytest.raises(ValueError, match="STQRY_API_URL"):
            Client()
```

**Step 2: Run to verify it fails**

```bash
cd python && pytest tests/test_client.py -v
```

Expected: `ImportError` or `AttributeError`.

**Step 3: Implement `python/stqry/__init__.py`**

```python
import os
from stqry._http import HttpClient
from stqry.collections import CollectionsResource
from stqry.screens import ScreensResource
from stqry.media import MediaResource
from stqry.codes import CodesResource
from stqry.projects import ProjectsResource


class Client:
    def __init__(self, api_url: str = None, token: str = None):
        api_url = api_url or os.environ.get("STQRY_API_URL")
        token = token or os.environ.get("STQRY_API_TOKEN")
        if not api_url or not token:
            raise ValueError(
                "api_url and token are required. Pass them explicitly or set "
                "STQRY_API_URL and STQRY_API_TOKEN environment variables."
            )
        self._http = HttpClient(api_url, token)
        self.collections = CollectionsResource(self._http)
        self.screens = ScreensResource(self._http)
        self.media = MediaResource(self._http)
        self.codes = CodesResource(self._http)
        self.projects = ProjectsResource(self._http)


__all__ = ["Client"]
```

At this point, create stub resource files so the import resolves. Each is just:

```python
# python/stqry/collections.py (stub)
class CollectionsResource:
    def __init__(self, http): self._http = http
```

Repeat for `screens.py`, `media.py`, `codes.py`, `projects.py`.

**Step 4: Run tests**

```bash
cd python && pytest tests/test_client.py -v
```

Expected: all 3 tests pass.

**Step 5: Commit**

```bash
git add python/stqry/
git commit -m "feat(python): add Client with env var auth"
```

---

### Task 4: Collections resource

**Files:**
- Modify: `python/stqry/collections.py`
- Create: `python/tests/test_collections.py`

**Step 1: Write the failing tests**

```python
# python/tests/test_collections.py
from unittest.mock import MagicMock
from stqry.collections import CollectionsResource


def _resource():
    return CollectionsResource(MagicMock())


def test_list():
    r = _resource()
    r._http.get.return_value = {"collections": [{"id": "1"}], "meta": {}}
    result = r.list(page=2)
    r._http.get.assert_called_once_with("/api/public/collections", params={"page": 2})
    assert result == [{"id": "1"}]


def test_get():
    r = _resource()
    r._http.get.return_value = {"collection": {"id": "1", "name": "Foo"}}
    result = r.get("1")
    r._http.get.assert_called_once_with("/api/public/collections/1", params=None)
    assert result == {"id": "1", "name": "Foo"}


def test_create():
    r = _resource()
    r._http.post.return_value = {"collection": {"id": "1"}}
    result = r.create(name="Foo")
    r._http.post.assert_called_once_with("/api/public/collections", {"collection": {"name": "Foo"}})
    assert result == {"id": "1"}


def test_update():
    r = _resource()
    r._http.patch.return_value = {"collection": {"id": "1", "name": "Bar"}}
    result = r.update("1", name="Bar")
    r._http.patch.assert_called_once_with("/api/public/collections/1", {"collection": {"name": "Bar"}})
    assert result == {"id": "1", "name": "Bar"}


def test_delete():
    r = _resource()
    r.delete("1")
    r._http.delete.assert_called_once_with("/api/public/collections/1", params=None)


def test_list_items():
    r = _resource()
    r._http.get.return_value = {"collection_items": [{"id": "a"}], "meta": {}}
    result = r.list_items("col1")
    r._http.get.assert_called_once_with("/api/public/collections/col1/collection_items", params={})
    assert result == [{"id": "a"}]


def test_create_item():
    r = _resource()
    r._http.post.return_value = {"collection_item": {"id": "a"}}
    result = r.create_item("col1", screen_id="s1")
    r._http.post.assert_called_once_with(
        "/api/public/collections/col1/collection_items",
        {"collection_item": {"screen_id": "s1"}},
    )
    assert result == {"id": "a"}


def test_reorder_items():
    r = _resource()
    r._http.post.return_value = {}
    r.reorder_items("col1", ["id2", "id1"])
    r._http.post.assert_called_once_with(
        "/api/public/collections/col1/collection_items/update_positions",
        {"ids": ["id2", "id1"]},
    )
```

**Step 2: Run to verify they fail**

```bash
cd python && pytest tests/test_collections.py -v
```

Expected: `AttributeError` — stub has no methods.

**Step 3: Implement `python/stqry/collections.py`**

```python
class CollectionsResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/collections", params=query or None)
        return resp.get("collections", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/collections/{id}", params=None)
        return resp.get("collection", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/collections", {"collection": fields})
        return resp.get("collection", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/collections/{id}", {"collection": fields})
        return resp.get("collection", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/collections/{id}", params=None)

    def list_items(self, collection_id: str, **query) -> list:
        path = f"/api/public/collections/{collection_id}/collection_items"
        resp = self._http.get(path, params=query)
        return resp.get("collection_items", [])

    def get_item(self, collection_id: str, item_id: str) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        resp = self._http.get(path, params=None)
        return resp.get("collection_item", resp)

    def create_item(self, collection_id: str, **fields) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items"
        resp = self._http.post(path, {"collection_item": fields})
        return resp.get("collection_item", resp)

    def update_item(self, collection_id: str, item_id: str, **fields) -> dict:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        resp = self._http.patch(path, {"collection_item": fields})
        return resp.get("collection_item", resp)

    def delete_item(self, collection_id: str, item_id: str) -> None:
        path = f"/api/public/collections/{collection_id}/collection_items/{item_id}"
        self._http.delete(path, params=None)

    def reorder_items(self, collection_id: str, item_ids: list) -> None:
        path = f"/api/public/collections/{collection_id}/collection_items/update_positions"
        self._http.post(path, {"ids": item_ids})
```

**Step 4: Run tests**

```bash
cd python && pytest tests/test_collections.py -v
```

Expected: all 8 tests pass.

**Step 5: Commit**

```bash
git add python/stqry/collections.py python/tests/test_collections.py
git commit -m "feat(python): add CollectionsResource"
```

---

### Task 5: Screens resource

**Files:**
- Modify: `python/stqry/screens.py`
- Create: `python/tests/test_screens.py`

**Step 1: Write the failing tests**

```python
# python/tests/test_screens.py
from unittest.mock import MagicMock
from stqry.screens import ScreensResource


def _resource():
    return ScreensResource(MagicMock())


def test_list():
    r = _resource()
    r._http.get.return_value = {"screens": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]
    r._http.get.assert_called_once_with("/api/public/screens", params=None)


def test_get():
    r = _resource()
    r._http.get.return_value = {"screen": {"id": "1"}}
    assert r.get("1") == {"id": "1"}


def test_create():
    r = _resource()
    r._http.post.return_value = {"screen": {"id": "1"}}
    result = r.create(name="Foo")
    r._http.post.assert_called_once_with("/api/public/screens", {"screen": {"name": "Foo"}})
    assert result == {"id": "1"}


def test_update():
    r = _resource()
    r._http.patch.return_value = {"screen": {"id": "1"}}
    r.update("1", name="Bar")
    r._http.patch.assert_called_once_with("/api/public/screens/1", {"screen": {"name": "Bar"}})


def test_delete():
    r = _resource()
    r.delete("1")
    r._http.delete.assert_called_once_with("/api/public/screens/1", params=None)


def test_list_sections():
    r = _resource()
    r._http.get.return_value = {"story_sections": [{"id": "s1"}], "meta": {}}
    result = r.list_sections("sc1")
    r._http.get.assert_called_once_with("/api/public/screens/sc1/story_sections", params={})
    assert result == [{"id": "s1"}]


def test_create_section():
    r = _resource()
    r._http.post.return_value = {"story_section": {"id": "s1"}}
    result = r.create_section("sc1", kind="text")
    r._http.post.assert_called_once_with(
        "/api/public/screens/sc1/story_sections", {"story_section": {"kind": "text"}}
    )
    assert result == {"id": "s1"}


def test_reorder_sections():
    r = _resource()
    r._http.post.return_value = {}
    r.reorder_sections("sc1", ["s2", "s1"])
    r._http.post.assert_called_once_with(
        "/api/public/screens/sc1/story_sections/update_positions",
        {"section_ids": ["s2", "s1"]},
    )


def test_list_sub_items():
    r = _resource()
    r._http.get.return_value = {"badge_items": [{"id": "b1"}]}
    result = r.list_sub_items("sc1", "s1", "badge_items")
    r._http.get.assert_called_once_with(
        "/api/public/screens/sc1/story_sections/s1/badge_items", params=None
    )
    assert result == [{"id": "b1"}]
```

**Step 2: Run to verify they fail**

```bash
cd python && pytest tests/test_screens.py -v
```

**Step 3: Implement `python/stqry/screens.py`**

```python
class ScreensResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/screens", params=query or None)
        return resp.get("screens", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/screens/{id}", params=None)
        return resp.get("screen", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/screens", {"screen": fields})
        return resp.get("screen", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/screens/{id}", {"screen": fields})
        return resp.get("screen", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/screens/{id}", params=None)

    # Story sections

    def list_sections(self, screen_id: str, **query) -> list:
        path = f"/api/public/screens/{screen_id}/story_sections"
        resp = self._http.get(path, params=query)
        return resp.get("story_sections", [])

    def get_section(self, screen_id: str, section_id: str) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        resp = self._http.get(path, params=None)
        return resp.get("story_section", resp)

    def create_section(self, screen_id: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections"
        resp = self._http.post(path, {"story_section": fields})
        return resp.get("story_section", resp)

    def update_section(self, screen_id: str, section_id: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        resp = self._http.patch(path, {"story_section": fields})
        return resp.get("story_section", resp)

    def delete_section(self, screen_id: str, section_id: str) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}"
        self._http.delete(path, params=None)

    def reorder_sections(self, screen_id: str, section_ids: list) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/update_positions"
        self._http.post(path, {"section_ids": section_ids})

    # Generic section sub-items

    def list_sub_items(self, screen_id: str, section_id: str, sub_item_type: str) -> list:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}"
        resp = self._http.get(path, params=None)
        return resp.get(sub_item_type, [])

    def create_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, singular_key: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}"
        resp = self._http.post(path, {singular_key: fields})
        return resp.get(singular_key, resp)

    def update_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, item_id: str, singular_key: str, **fields) -> dict:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}/{item_id}"
        resp = self._http.patch(path, {singular_key: fields})
        return resp.get(singular_key, resp)

    def delete_sub_item(self, screen_id: str, section_id: str, sub_item_type: str, item_id: str) -> None:
        path = f"/api/public/screens/{screen_id}/story_sections/{section_id}/{sub_item_type}/{item_id}"
        self._http.delete(path, params=None)
```

**Step 4: Run tests**

```bash
cd python && pytest tests/test_screens.py -v
```

Expected: all 9 tests pass.

**Step 5: Commit**

```bash
git add python/stqry/screens.py python/tests/test_screens.py
git commit -m "feat(python): add ScreensResource"
```

---

### Task 6: Media, Codes, and Projects resources

**Files:**
- Modify: `python/stqry/media.py`
- Modify: `python/stqry/codes.py`
- Modify: `python/stqry/projects.py`
- Create: `python/tests/test_media.py`
- Create: `python/tests/test_codes.py`
- Create: `python/tests/test_projects.py`

**Step 1: Write the failing tests**

```python
# python/tests/test_media.py
from unittest.mock import MagicMock
from stqry.media import MediaResource

def _r(): return MediaResource(MagicMock())

def test_list():
    r = _r()
    r._http.get.return_value = {"media_items": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]

def test_get():
    r = _r()
    r._http.get.return_value = {"media_item": {"id": "1"}}
    assert r.get("1") == {"id": "1"}

def test_create():
    r = _r()
    r._http.post.return_value = {"media_item": {"id": "1"}}
    assert r.create(title="Foo") == {"id": "1"}

def test_update():
    r = _r()
    r._http.patch.return_value = {"media_item": {"id": "1"}}
    r.update("1", title="Bar")
    r._http.patch.assert_called_once_with("/api/public/media_items/1", {"media_item": {"title": "Bar"}})

def test_delete():
    r = _r()
    r.delete("1")
    r._http.delete.assert_called_once_with("/api/public/media_items/1", params=None)

def test_delete_with_language():
    r = _r()
    r.delete("1", language="fr")
    r._http.delete.assert_called_once_with("/api/public/media_items/1", params={"language": "fr"})
```

```python
# python/tests/test_codes.py
from unittest.mock import MagicMock
from stqry.codes import CodesResource

def _r(): return CodesResource(MagicMock())

def test_list():
    r = _r()
    r._http.get.return_value = {"codes": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]

def test_get():
    r = _r()
    r._http.get.return_value = {"code": {"id": "1"}}
    assert r.get("1") == {"id": "1"}

def test_create():
    r = _r()
    r._http.post.return_value = {"code": {"id": "1"}}
    assert r.create(value="ABC") == {"id": "1"}

def test_update():
    r = _r()
    r._http.patch.return_value = {"code": {"id": "1"}}
    r.update("1", value="XYZ")
    r._http.patch.assert_called_once_with("/api/public/codes/1", {"code": {"value": "XYZ"}})

def test_delete():
    r = _r()
    r.delete("1")
    r._http.delete.assert_called_once_with("/api/public/codes/1", params=None)
```

```python
# python/tests/test_projects.py
from unittest.mock import MagicMock
from stqry.projects import ProjectsResource

def _r(): return ProjectsResource(MagicMock())

def test_list():
    r = _r()
    r._http.get.return_value = {"projects": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]

def test_get():
    r = _r()
    r._http.get.return_value = {"project": {"id": "1"}}
    assert r.get("1") == {"id": "1"}
```

**Step 2: Run to verify they fail**

```bash
cd python && pytest tests/test_media.py tests/test_codes.py tests/test_projects.py -v
```

**Step 3: Implement the three resource files**

```python
# python/stqry/media.py
class MediaResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/media_items", params=query or None)
        return resp.get("media_items", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/media_items/{id}", params=None)
        return resp.get("media_item", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/media_items", {"media_item": fields})
        return resp.get("media_item", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/media_items/{id}", {"media_item": fields})
        return resp.get("media_item", resp)

    def delete(self, id: str, **query) -> None:
        self._http.delete(f"/api/public/media_items/{id}", params=query or None)
```

```python
# python/stqry/codes.py
class CodesResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/codes", params=query or None)
        return resp.get("codes", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/codes/{id}", params=None)
        return resp.get("code", resp)

    def create(self, **fields) -> dict:
        resp = self._http.post("/api/public/codes", {"code": fields})
        return resp.get("code", resp)

    def update(self, id: str, **fields) -> dict:
        resp = self._http.patch(f"/api/public/codes/{id}", {"code": fields})
        return resp.get("code", resp)

    def delete(self, id: str) -> None:
        self._http.delete(f"/api/public/codes/{id}", params=None)
```

```python
# python/stqry/projects.py
class ProjectsResource:
    def __init__(self, http):
        self._http = http

    def list(self, **query) -> list:
        resp = self._http.get("/api/public/projects", params=query or None)
        return resp.get("projects", [])

    def get(self, id: str) -> dict:
        resp = self._http.get(f"/api/public/projects/{id}", params=None)
        return resp.get("project", resp)
```

**Step 4: Run tests**

```bash
cd python && pytest tests/ -v
```

Expected: all tests pass.

**Step 5: Commit**

```bash
git add python/stqry/media.py python/stqry/codes.py python/stqry/projects.py python/tests/
git commit -m "feat(python): add Media, Codes, Projects resources"
```

---

### Task 7: CI — PyPI publishing

**Files:**
- Modify: `.github/workflows/release.yml`

**Step 1: Add `publish-pypi` job**

Append to `.github/workflows/release.yml` after the existing `release` job:

```yaml
  publish-pypi:
    needs: [release]
    runs-on: ubuntu-latest
    environment: release
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-python@v5
        with:
          python-version: "3.12"

      - name: Build package
        run: |
          pip install build twine
          python -m build
        working-directory: python/

      - name: Publish to PyPI
        run: twine upload dist/*
        working-directory: python/
        env:
          TWINE_USERNAME: __token__
          TWINE_PASSWORD: ${{ secrets.PYPI_TOKEN }}
```

**Step 2: Add `PYPI_TOKEN` secret**

In GitHub: Settings → Environments → `release` → Add secret `PYPI_TOKEN`.

Get the token from pypi.org → Account Settings → API tokens → Add API token (scope: entire account for first publish, then narrow to the `stqry` project).

**Step 3: Verify workflow YAML is valid**

```bash
cd /path/to/repo && cat .github/workflows/release.yml
```

Check indentation — the `publish-pypi` job must be at the same level as `release` (under `jobs:`).

**Step 4: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: publish Python SDK to PyPI on release"
```

---

### Task 8: Full test run + smoke test

**Step 1: Run all tests**

```bash
cd python && pytest tests/ -v
```

Expected: all tests pass, no warnings.

**Step 2: Smoke test the build locally**

```bash
cd python && pip install build && python -m build
ls dist/
```

Expected: `stqry-*.whl` and `stqry-*.tar.gz` present.

Note: Version will show as `0.1.dev0` locally (no git tag). That's fine — CI will build from the tag.

**Step 3: Final commit if needed, then push**

```bash
git push origin main
```

---

## Post-implementation

- Add `PYPI_TOKEN` secret to the `release` GitHub environment before the next tag push
- First PyPI publish requires the package name `stqry` to be available — check at pypi.org/project/stqry before tagging
- After first successful publish, narrow the PyPI token scope to just the `stqry` project
