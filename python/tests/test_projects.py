from unittest.mock import MagicMock
from stqry.projects import ProjectsResource


def _r():
    return ProjectsResource(MagicMock())


def test_list():
    r = _r()
    r._http.get.return_value = {"projects": [{"id": "1"}], "meta": {}}
    assert r.list() == [{"id": "1"}]
    r._http.get.assert_called_once_with("/api/public/projects", params=None)


def test_get():
    r = _r()
    r._http.get.return_value = {"project": {"id": "1"}}
    assert r.get("1") == {"id": "1"}
    r._http.get.assert_called_once_with("/api/public/projects/1", params=None)
