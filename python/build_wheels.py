#!/usr/bin/env python3
"""Build platform-specific wheels from GoReleaser release artifacts."""

import base64
import hashlib
import shutil
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


def _record_entry(rel_path: str, abs_path: Path) -> str:
    digest = hashlib.sha256(abs_path.read_bytes()).digest()
    h = "sha256=" + base64.urlsafe_b64encode(digest).rstrip(b"=").decode()
    size = abs_path.stat().st_size
    return f"{rel_path},{h},{size}"


_RUN_PY = """\
import os
import subprocess
import sys


def main():
    binary = os.path.join(os.path.dirname(os.path.abspath(__file__)), "bin", "stqry")
    if sys.platform == "win32":
        binary += ".exe"
    sys.exit(subprocess.call([binary] + sys.argv[1:]))
"""


def extract_binary(archive_path: Path, go_os: str, dest_dir: Path) -> Path:
    name = binary_name(go_os)
    dest = dest_dir / name
    if go_os == "windows":
        with zipfile.ZipFile(archive_path) as zf:
            zf.extract(name, dest_dir)
    else:
        with tarfile.open(archive_path) as tf:
            member = tf.getmember(name)
            member.name = name
            tf.extract(member, dest_dir)
    if go_os != "windows":
        dest.chmod(0o755)
    return dest


def build_wheel(
    binary_path: Path,
    go_os: str,
    go_arch: str,
    version: str,
    output_dir: Path,
) -> Path:
    tag = platform_tag(go_os, go_arch)
    wheel_name = f"stqry-{version}-py3-none-{tag}.whl"
    output_dir.mkdir(parents=True, exist_ok=True)
    wheel_path = output_dir / wheel_name

    with tempfile.TemporaryDirectory() as tmp:
        tmp = Path(tmp)

        pkg = tmp / "stqry"
        bin_dir = pkg / "bin"
        bin_dir.mkdir(parents=True)

        dest_binary = bin_dir / binary_name(go_os)
        shutil.copy2(binary_path, dest_binary)
        if go_os != "windows":
            dest_binary.chmod(0o755)

        (pkg / "__init__.py").write_text(f'__version__ = "{version}"\n')
        (pkg / "_run.py").write_text(_RUN_PY)

        dist_info = tmp / f"stqry-{version}.dist-info"
        dist_info.mkdir()
        (dist_info / "METADATA").write_text(
            f"Metadata-Version: 2.1\n"
            f"Name: stqry\n"
            f"Version: {version}\n"
            f"Summary: STQRY CLI - manage collections, screens, media, and content\n"
            f"Home-page: https://github.com/mytours/stqry-cli\n"
            f"License: MIT\n"
            f"Requires-Python: >=3.8\n"
        )
        (dist_info / "WHEEL").write_text(
            f"Wheel-Version: 1.0\n"
            f"Generator: build_wheels.py\n"
            f"Root-Is-Purelib: false\n"
            f"Tag: py3-none-{tag}\n"
        )
        (dist_info / "entry_points.txt").write_text(
            "[console_scripts]\nstqry = stqry._run:main\n"
        )

        # Collect all files for RECORD
        all_files = [f for f in tmp.rglob("*") if f.is_file()]
        record_path = dist_info / "RECORD"
        record_lines = [
            _record_entry(str(f.relative_to(tmp)), f) for f in all_files
        ]
        # RECORD's own entry has empty hash and size
        record_lines.append(f"{dist_info.name}/RECORD,,")
        record_path.write_text("\n".join(record_lines) + "\n")

        with zipfile.ZipFile(wheel_path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
            for f in tmp.rglob("*"):
                if f.is_file():
                    zf.write(f, f.relative_to(tmp))

    return wheel_path
