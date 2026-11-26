#!/usr/bin/env python3
"""Local development setup helper for MET (Windows x86_64).

Fetches OCR models from MaaXYZ/MaaCommonAssets and optionally downloads
Windows x86_64 releases of MaaFramework and MXU into deps/.
"""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
import tempfile
import urllib.request
import zipfile
from pathlib import Path

ROOT_DIR = Path(__file__).resolve().parent.parent
ASSETS_DIR = ROOT_DIR / "assets"
OCR_TARGET_DIR = ASSETS_DIR / "resource" / "base" / "model" / "ocr"
COMMON_ASSETS_DIR = ASSETS_DIR / "MaaCommonAssets"
DEPS_DIR = ROOT_DIR / "deps"

OCR_SUBPATH = Path("OCR") / "ppocr_v6" / "small"


def configure_ocr_model(force: bool = False) -> None:
    if OCR_TARGET_DIR.exists() and any(OCR_TARGET_DIR.iterdir()) and not force:
        print(f"OCR model already exists at {OCR_TARGET_DIR}, skipping.")
        return

    local_src = COMMON_ASSETS_DIR / OCR_SUBPATH
    if local_src.exists():
        print(f"Copying OCR model from local {local_src}...")
        shutil.copytree(local_src, OCR_TARGET_DIR, dirs_exist_ok=True)
        return

    print("Cloning OCR model (sparse checkout) from MaaXYZ/MaaCommonAssets...")
    with tempfile.TemporaryDirectory() as tmp:
        tmp_repo = Path(tmp) / "MaaCommonAssets"
        subprocess.run(
            [
                "git",
                "clone",
                "--depth=1",
                "--filter=blob:none",
                "--sparse",
                "https://github.com/MaaXYZ/MaaCommonAssets.git",
                str(tmp_repo),
            ],
            check=True,
        )
        subprocess.run(
            ["git", "-C", str(tmp_repo), "sparse-checkout", "set", "OCR/ppocr_v6/small"],
            check=True,
        )
        src = tmp_repo / OCR_SUBPATH
        if not src.exists():
            raise FileNotFoundError(f"Expected OCR directory not found: {src}")
        shutil.copytree(src, OCR_TARGET_DIR, dirs_exist_ok=True)
    print(f"OCR model configured at {OCR_TARGET_DIR}")


def fetch_latest_tag(repo: str) -> str:
    url = f"https://github.com/{repo}/releases/latest"
    req = urllib.request.Request(url, headers={"User-Agent": "MET-Configure"})
    with urllib.request.urlopen(req, timeout=30) as resp:
        final_url = resp.geturl()
    return final_url.rstrip("/").split("/")[-1]


def download_and_extract_zip(url: str, dest: Path) -> None:
    print(f"Downloading {url}...")
    with tempfile.TemporaryDirectory() as tmp:
        zip_path = Path(tmp) / "archive.zip"
        req = urllib.request.Request(url, headers={"User-Agent": "MET-Configure"})
        with urllib.request.urlopen(req, timeout=120) as resp, zip_path.open("wb") as out:
            shutil.copyfileobj(resp, out)
        if dest.exists():
            shutil.rmtree(dest)
        dest.mkdir(parents=True, exist_ok=True)
        with zipfile.ZipFile(zip_path, "r") as zf:
            zf.extractall(dest)
    print(f"Extracted to {dest}")


def download_windows_runtime(maafw_version: str = "", mxu_version: str = "") -> None:
    maafw_tag = maafw_version or fetch_latest_tag("MaaXYZ/MaaFramework")
    mxu_tag = mxu_version or fetch_latest_tag("MaaXYZ/MXU")

    maafw_url = (
        f"https://github.com/MaaXYZ/MaaFramework/releases/download/"
        f"{maafw_tag}/MAA-win-x86_64-{maafw_tag}.zip"
    )
    mxu_url = (
        f"https://github.com/MaaXYZ/MXU/releases/download/"
        f"{mxu_tag}/MXU-win-x86_64-{mxu_tag}.zip"
    )

    download_and_extract_zip(maafw_url, DEPS_DIR / "maafw")
    download_and_extract_zip(mxu_url, DEPS_DIR / "mxu")


def main() -> int:
    parser = argparse.ArgumentParser(description="Configure MET local dev dependencies.")
    parser.add_argument("--force-ocr", action="store_true", help="Re-copy OCR model even if present.")
    parser.add_argument(
        "--download-runtime",
        action="store_true",
        help="Download Windows x86_64 MaaFramework and MXU into deps/.",
    )
    parser.add_argument("--maafw-version", default="", help="Pin MaaFramework tag (e.g. v5.16.1).")
    parser.add_argument("--mxu-version", default="", help="Pin MXU tag (e.g. v2.11.2).")
    args = parser.parse_args()

    configure_ocr_model(force=args.force_ocr)
    if args.download_runtime:
        download_windows_runtime(args.maafw_version, args.mxu_version)
    return 0


if __name__ == "__main__":
    sys.exit(main())
