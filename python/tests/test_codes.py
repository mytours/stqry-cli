from unittest.mock import MagicMock
from stqry.codes import CodesResource


def _r():
    return CodesResource(MagicMock())


def test_list():
    r = _r()
    r._http.get.return_value = {"codes": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]
    r._http.get.assert_called_once_with("/api/public/codes", params=None)


def test_get():
    r = _r()
    r._http.get.return_value = {"code": {"id": "1"}}
    assert r.get("1") == {"id": "1"}
    r._http.get.assert_called_once_with("/api/public/codes/1", params=None)


def test_create():
    r = _r()
    r._http.post.return_value = {"code": {"id": "1"}}
    assert r.create(value="ABC") == {"id": "1"}
    r._http.post.assert_called_once_with("/api/public/codes", {"code": {"value": "ABC"}})


def test_update():
    r = _r()
    r._http.patch.return_value = {"code": {"id": "1"}}
    r.update("1", value="XYZ")
    r._http.patch.assert_called_once_with("/api/public/codes/1", {"code": {"value": "XYZ"}})


def test_delete():
    r = _r()
    r.delete("1")
    r._http.delete.assert_called_once_with("/api/public/codes/1", params=None)
