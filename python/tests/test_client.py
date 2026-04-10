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
