from unittest.mock import MagicMock
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
