import io
import stat
import tarfile
import unittest.mock as mock
import zipfile as _zipfile
from pathlib import Path

import pytest
import build_wheels as bw


def test_platform_tag_darwin_arm64():
    assert bw.platform_tag("darwin", "arm64") == "macosx_11_0_arm64"

def test_platform_tag_darwin_amd64():
    assert bw.platform_tag("darwin", "amd64") == "macosx_10_9_x86_64"

def test_platform_tag_linux_arm64():
    assert bw.platform_tag("linux", "arm64") == "manylinux_2_17_aarch64.manylinux2014_aarch64"

def test_platform_tag_linux_amd64():
    assert bw.platform_tag("linux", "amd64") == "manylinux_2_17_x86_64.manylinux2014_x86_64"

def test_platform_tag_windows_amd64():
    assert bw.platform_tag("windows", "amd64") == "win_amd64"

def test_platform_tag_windows_arm64():
    assert bw.platform_tag("windows", "arm64") == "win_arm64"

def test_archive_name_unix():
    assert bw.archive_name("darwin", "arm64") == "stqry-cli_darwin_arm64.tar.gz"
    assert bw.archive_name("linux", "amd64") == "stqry-cli_linux_amd64.tar.gz"

def test_archive_name_windows():
    assert bw.archive_name("windows", "amd64") == "stqry-cli_windows_amd64.zip"

def test_binary_name_windows():
    assert bw.binary_name("windows") == "stqry.exe"

def test_binary_name_unix():
    assert bw.binary_name("darwin") == "stqry"
    assert bw.binary_name("linux") == "stqry"


def test_extract_binary_from_tar(tmp_path):
    # Create a fake tar.gz with a stqry binary inside
    binary_content = b"fake elf binary"
    archive = tmp_path / "stqry_darwin_arm64.tar.gz"
    with tarfile.open(archive, "w:gz") as tf:
        data = io.BytesIO(binary_content)
        info = tarfile.TarInfo(name="stqry")
        info.size = len(binary_content)
        tf.addfile(info, data)

    dest = tmp_path / "out"
    dest.mkdir()
    result = bw.extract_binary(archive, "darwin", dest)

    assert result.name == "stqry"
    assert result.read_bytes() == binary_content
    assert result.stat().st_mode & stat.S_IXUSR  # executable bit set


def test_extract_binary_from_zip(tmp_path):
    binary_content = b"fake pe binary"
    archive = tmp_path / "stqry_windows_amd64.zip"
    with _zipfile.ZipFile(archive, "w") as zf:
        zf.writestr("stqry.exe", binary_content)

    dest = tmp_path / "out"
    dest.mkdir()
    result = bw.extract_binary(archive, "windows", dest)

    assert result.name == "stqry.exe"
    assert result.read_bytes() == binary_content


def test_build_wheel_filename(tmp_path):
    binary = tmp_path / "stqry"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "darwin", "arm64", "0.6.3", out)

    assert wheel.name == "stqry-0.6.3-py3-none-macosx_11_0_arm64.whl"


def test_build_wheel_contains_expected_files(tmp_path):
    binary = tmp_path / "stqry"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "linux", "amd64", "0.6.3", out)

    with _zipfile.ZipFile(wheel) as zf:
        names = zf.namelist()

    assert "stqry/__init__.py" in names
    assert "stqry/_run.py" in names
    assert "stqry/bin/stqry" in names
    assert "stqry-0.6.3.dist-info/METADATA" in names
    assert "stqry-0.6.3.dist-info/WHEEL" in names
    assert "stqry-0.6.3.dist-info/entry_points.txt" in names
    assert "stqry-0.6.3.dist-info/RECORD" in names


def test_build_wheel_windows_uses_exe(tmp_path):
    binary = tmp_path / "stqry.exe"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "windows", "amd64", "0.6.3", out)

    with _zipfile.ZipFile(wheel) as zf:
        names = zf.namelist()

    assert "stqry/bin/stqry.exe" in names
    assert "stqry/bin/stqry" not in names


def test_build_wheel_metadata_tag(tmp_path):
    binary = tmp_path / "stqry"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "linux", "arm64", "0.6.3", out)

    with _zipfile.ZipFile(wheel) as zf:
        wheel_meta = zf.read("stqry-0.6.3.dist-info/WHEEL").decode()

    assert "Tag: py3-none-manylinux_2_17_aarch64.manylinux2014_aarch64" in wheel_meta


def test_build_wheel_entry_points(tmp_path):
    binary = tmp_path / "stqry"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "darwin", "amd64", "0.6.3", out)

    with _zipfile.ZipFile(wheel) as zf:
        ep = zf.read("stqry-0.6.3.dist-info/entry_points.txt").decode()

    assert "stqry = stqry._run:main" in ep


def test_build_wheel_run_py_matches_source(tmp_path):
    """_run.py in the wheel must match the on-disk source to prevent drift."""
    binary = tmp_path / "stqry"
    binary.write_bytes(b"fake")

    out = tmp_path / "dist"
    wheel = bw.build_wheel(binary, "linux", "amd64", "0.6.3", out)

    source_path = Path(bw.__file__).parent / "stqry" / "_run.py"
    expected = source_path.read_text()

    with _zipfile.ZipFile(wheel) as zf:
        actual = zf.read("stqry/_run.py").decode()

    assert actual == expected


def test_download_artifact_failure(tmp_path):
    failed = mock.Mock()
    failed.returncode = 1
    failed.stderr = "release not found"

    with mock.patch("subprocess.run", return_value=failed):
        with pytest.raises(RuntimeError, match="gh release download failed"):
            bw.download_artifact("v1.0.0", "stqry-cli_darwin_arm64.tar.gz", tmp_path)
