from unittest.mock import MagicMock
from stqry.media import MediaResource


def _r():
    return MediaResource(MagicMock())


def test_list():
    r = _r()
    r._http.get.return_value = {"media_items": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]
    r._http.get.assert_called_once_with("/api/public/media_items", params=None)


def test_get():
    r = _r()
    r._http.get.return_value = {"media_item": {"id": "1"}}
    assert r.get("1") == {"id": "1"}
    r._http.get.assert_called_once_with("/api/public/media_items/1", params=None)


def test_create():
    r = _r()
    r._http.post.return_value = {"media_item": {"id": "1"}}
    assert r.create(title="Foo") == {"id": "1"}
    r._http.post.assert_called_once_with("/api/public/media_items", {"media_item": {"title": "Foo"}})


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
