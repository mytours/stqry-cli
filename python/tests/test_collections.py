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
    r._http.get.assert_called_once_with("/api/public/collections/col1/collection_items", params=None)
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
