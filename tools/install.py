#!/usr/bin/env python3
"""Assemble the Windows x86_64 MXU distribution directory for MET.

Output layout (install/):
  install/
  ├── mxu.exe
  ├── maafw/           # MaaFramework bin/ contents
  ├── interface.json   # PI V2 config with injected version
  ├── resource/        # base/ + bilibili/ resources (including OCR models)
  └── README.md
"""

from __future__ import annotations

import argparse
import json
import shutil
import sys
from pathlib import Path

ROOT_DIR = Path(__file__).resolve().parent.parent
ASSETS_DIR = ROOT_DIR / "assets"
DEPS_DIR = ROOT_DIR / "deps"
INSTALL_DIR = ROOT_DIR / "install"


def find_mxu_exe(mxu_dir: Path) -> Path | None:
    if not mxu_dir.exists():
        return None
    for candidate in mxu_dir.rglob("mxu.exe"):
        return candidate
    return None


def copy_resources(allow_missing_ocr: bool) -> None:
    src_resource = ASSETS_DIR / "resource"
    if not src_resource.exists():
        raise FileNotFoundError(f"Missing assets/resource at {src_resource}")

    ocr_dir = src_resource / "base" / "model" / "ocr"
    if not allow_missing_ocr and (not ocr_dir.exists() or not any(ocr_dir.iterdir())):
        raise FileNotFoundError(
            f"OCR model directory is missing or empty: {ocr_dir}\n"
            "Run 'python tools/configure.py' first or pass --allow-missing-ocr for dry runs."
        )

    shutil.copytree(
        src_resource,
        INSTALL_DIR / "resource",
        ignore=shutil.ignore_patterns(".gitkeep"),
        dirs_exist_ok=True,
    )


def install_interface(version: str) -> None:
    src_interface = ASSETS_DIR / "interface.json"
    with src_interface.open("r", encoding="utf-8") as f:
        data = json.load(f)

    data["version"] = version
    if "title" in data:
        data["title"] = f"MET - 悠久之树助手 ({version})"

    dst_interface = INSTALL_DIR / "interface.json"
    with dst_interface.open("w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=4)
        f.write("\n")


def install_runtime(allow_missing_runtime: bool) -> None:
    maafw_bin = DEPS_DIR / "maafw" / "bin"
    mxu_exe = find_mxu_exe(DEPS_DIR / "mxu")

    if maafw_bin.exists():
        shutil.copytree(maafw_bin, INSTALL_DIR / "maafw", dirs_exist_ok=True)
    elif not allow_missing_runtime:
        raise FileNotFoundError(
            f"Missing MaaFramework runtime at {maafw_bin}. "
            "Run 'python tools/configure.py --download-runtime' first."
        )

    if mxu_exe is not None:
        shutil.copy2(mxu_exe, INSTALL_DIR / "mxu.exe")
    elif not allow_missing_runtime:
        raise FileNotFoundError(
            f"Missing mxu.exe under {DEPS_DIR / 'mxu'}. "
            "Run 'python tools/configure.py --download-runtime' first."
        )


def main() -> int:
    parser = argparse.ArgumentParser(description="Assemble MET Windows x86_64 package.")
    parser.add_argument("version", nargs="?", default="v0.0.0-dev", help="Version tag to inject.")
    parser.add_argument(
        "--allow-missing-runtime",
        action="store_true",
        help="Skip error if deps/maafw or deps/mxu are missing (local dry run).",
    )
    parser.add_argument(
        "--allow-missing-ocr",
        action="store_true",
        help="Skip error if OCR models are missing (local dry run).",
    )
    args = parser.parse_args()

    if INSTALL_DIR.exists():
        shutil.rmtree(INSTALL_DIR)
    INSTALL_DIR.mkdir(parents=True, exist_ok=True)

    install_runtime(allow_missing_runtime=args.allow_missing_runtime)
    copy_resources(allow_missing_ocr=args.allow_missing_ocr)
    install_interface(args.version)

    readme = ROOT_DIR / "README.md"
    if readme.exists():
        shutil.copy2(readme, INSTALL_DIR / "README.md")

    logo = ASSETS_DIR / "logo.png"
    if logo.exists():
        shutil.copy2(logo, INSTALL_DIR / "logo.png")

    print(f"Assembled MET {args.version} into {INSTALL_DIR}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
