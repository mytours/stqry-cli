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
    assert bw.archive_name("darwin", "arm64") == "stqry_darwin_arm64.tar.gz"
    assert bw.archive_name("linux", "amd64") == "stqry_linux_amd64.tar.gz"

def test_archive_name_windows():
    assert bw.archive_name("windows", "amd64") == "stqry_windows_amd64.zip"

def test_binary_name_windows():
    assert bw.binary_name("windows") == "stqry.exe"

def test_binary_name_unix():
    assert bw.binary_name("darwin") == "stqry"
    assert bw.binary_name("linux") == "stqry"
