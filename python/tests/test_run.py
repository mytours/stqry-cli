import sys
import importlib
from unittest.mock import patch
import pytest


def test_main_invokes_binary_with_args(monkeypatch, tmp_path):
    # Arrange: point __file__ at a temp dir so dirname resolves predictably
    fake_bin = tmp_path / "bin" / "stqry"
    fake_bin.parent.mkdir()
    fake_bin.write_bytes(b"fake")

    monkeypatch.setattr(sys, "argv", ["stqry", "collections", "list"])

    import stqry._run as run_module
    monkeypatch.setattr(run_module, "__file__", str(tmp_path / "_run.py"))

    with patch("subprocess.call", return_value=0) as mock_call:
        with pytest.raises(SystemExit) as exc:
            run_module.main()

    assert exc.value.code == 0
    assert mock_call.call_args[0][0] == [str(fake_bin), "collections", "list"]


def test_main_propagates_nonzero_exit(monkeypatch, tmp_path):
    fake_bin = tmp_path / "bin" / "stqry"
    fake_bin.parent.mkdir()
    fake_bin.write_bytes(b"fake")

    monkeypatch.setattr(sys, "argv", ["stqry", "--version"])

    import stqry._run as run_module
    monkeypatch.setattr(run_module, "__file__", str(tmp_path / "_run.py"))

    with patch("subprocess.call", return_value=2):
        with pytest.raises(SystemExit) as exc:
            run_module.main()

    assert exc.value.code == 2
