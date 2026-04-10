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
    r._http.get.assert_called_once_with("/api/public/screens/sc1/story_sections", params=None)
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
