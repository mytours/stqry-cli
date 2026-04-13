#!/usr/bin/env python3
"""Build platform-specific wheels from GoReleaser release artifacts."""

import argparse
import shutil
import subprocess
import sys
import tarfile
import tempfile
import zipfile
from pathlib import Path

PLATFORMS = [
    ("darwin", "arm64"),
    ("darwin", "amd64"),
    ("linux", "arm64"),
    ("linux", "amd64"),
    ("windows", "amd64"),
    ("windows", "arm64"),
]

_PLATFORM_TAGS = {
    ("darwin", "arm64"): "macosx_11_0_arm64",
    ("darwin", "amd64"): "macosx_10_9_x86_64",
    ("linux", "arm64"): "manylinux_2_17_aarch64.manylinux2014_aarch64",
    ("linux", "amd64"): "manylinux_2_17_x86_64.manylinux2014_x86_64",
    ("windows", "amd64"): "win_amd64",
    ("windows", "arm64"): "win_arm64",
}


def platform_tag(go_os: str, go_arch: str) -> str:
    return _PLATFORM_TAGS[(go_os, go_arch)]


def archive_name(go_os: str, go_arch: str) -> str:
    ext = "zip" if go_os == "windows" else "tar.gz"
    return f"stqry_{go_os}_{go_arch}.{ext}"


def binary_name(go_os: str) -> str:
    return "stqry.exe" if go_os == "windows" else "stqry"
