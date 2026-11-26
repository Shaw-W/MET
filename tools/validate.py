#!/usr/bin/env python3
"""Static validator for MET (Maa Eternal Tree) PI V2 interface & Pipeline JSON files."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

ROOT_DIR = Path(__file__).resolve().parent.parent
ASSETS_DIR = ROOT_DIR / "assets"
INTERFACE_PATH = ASSETS_DIR / "interface.json"
BASE_PIPELINE_DIR = ASSETS_DIR / "resource" / "base" / "pipeline"
BILIBILI_PIPELINE_DIR = ASSETS_DIR / "resource" / "bilibili" / "pipeline"

FORBIDDEN_TOP_KEYS = {
    "mirrorchyan_rid",
    "mirrorchyan_multiplatform",
    "custom_title",
}

NODE_PREFIX_RE = re.compile(r"^(?:\[[^\]]+\])+")


def strip_node_prefix(name: str) -> str:
    return NODE_PREFIX_RE.sub("", name)


def _validate_override_map(
    owner: str,
    override_map: object,
    known_nodes: set[str],
    errors: list[str],
) -> None:
    if override_map is None:
        return
    if not isinstance(override_map, dict):
        errors.append(f"assets/interface.json: {owner} pipeline_override must be an object")
        return
    for node_name, node_body in override_map.items():
        if node_name not in known_nodes:
            errors.append(
                f"assets/interface.json: {owner} overrides unknown node '{node_name}'"
            )
        if not isinstance(node_body, dict):
            continue
        for field in ("next", "on_error", "interrupt"):
            refs = node_body.get(field)
            if refs is None:
                continue
            if isinstance(refs, str):
                refs = [refs]
            if not isinstance(refs, list):
                continue
            for raw_ref in refs:
                if not isinstance(raw_ref, str):
                    continue
                target = strip_node_prefix(raw_ref)
                if target and target not in known_nodes:
                    errors.append(
                        f"assets/interface.json: {owner} '{node_name}.{field}' references unknown node '{target}'"
                    )


def load_json(path: Path, errors: list[str]) -> dict | None:
    try:
        with path.open("r", encoding="utf-8") as f:
            data = json.load(f)
        if not isinstance(data, dict):
            errors.append(f"{path.relative_to(ROOT_DIR)}: root must be a JSON object")
            return None
        return data
    except Exception as exc:
        errors.append(f"{path.relative_to(ROOT_DIR)}: invalid JSON ({exc})")
        return None


def collect_pipelines(pipeline_dir: Path, errors: list[str]) -> dict[str, Path]:
    nodes: dict[str, Path] = {}
    if not pipeline_dir.exists():
        errors.append(f"Missing pipeline directory: {pipeline_dir.relative_to(ROOT_DIR)}")
        return nodes

    for file_path in sorted(pipeline_dir.rglob("*.json")):
        data = load_json(file_path, errors)
        if data is None:
            continue
        for node_name, node_body in data.items():
            if node_name in nodes:
                errors.append(
                    f"Duplicate node '{node_name}' in {file_path.relative_to(ROOT_DIR)} "
                    f"(already defined in {nodes[node_name].relative_to(ROOT_DIR)})"
                )
            else:
                nodes[node_name] = file_path
            if not isinstance(node_body, dict):
                errors.append(
                    f"{file_path.relative_to(ROOT_DIR)}: node '{node_name}' must be an object"
                )
    return nodes


def validate_node_references(
    pipeline_dir: Path, known_nodes: set[str], errors: list[str]
) -> None:
    if not pipeline_dir.exists():
        return
    for file_path in sorted(pipeline_dir.rglob("*.json")):
        data = load_json(file_path, errors)
        if data is None:
            continue
        for node_name, node_body in data.items():
            if not isinstance(node_body, dict):
                continue
            for field in ("next", "on_error", "interrupt"):
                refs = node_body.get(field)
                if refs is None:
                    continue
                if isinstance(refs, str):
                    refs = [refs]
                if not isinstance(refs, list):
                    errors.append(
                        f"{file_path.relative_to(ROOT_DIR)}: '{node_name}.{field}' must be string or list"
                    )
                    continue
                for raw_ref in refs:
                    if not isinstance(raw_ref, str):
                        continue
                    target = strip_node_prefix(raw_ref)
                    if target and target not in known_nodes:
                        errors.append(
                            f"{file_path.relative_to(ROOT_DIR)}: '{node_name}.{field}' references unknown node '{target}'"
                        )


def main() -> int:
    errors: list[str] = []

    interface = load_json(INTERFACE_PATH, errors)
    base_nodes = collect_pipelines(BASE_PIPELINE_DIR, errors)
    bili_nodes = (
        collect_pipelines(BILIBILI_PIPELINE_DIR, errors)
        if BILIBILI_PIPELINE_DIR.exists()
        else {}
    )

    all_base_names = set(base_nodes.keys())
    all_known_names = all_base_names | set(bili_nodes.keys())

    validate_node_references(BASE_PIPELINE_DIR, all_base_names, errors)
    validate_node_references(BILIBILI_PIPELINE_DIR, all_known_names, errors)

    if interface is not None:
        if interface.get("interface_version") != 2:
            errors.append("assets/interface.json: 'interface_version' must be 2")

        for key in FORBIDDEN_TOP_KEYS:
            if key in interface:
                errors.append(
                    f"assets/interface.json: forbidden key '{key}' found (violates SPEC.md)"
                )

        for req in ("name", "controller", "resource", "task"):
            if req not in interface:
                errors.append(f"assets/interface.json: missing required field '{req}'")

        defined_options = set((interface.get("option") or {}).keys())
        for task in interface.get("task") or []:
            if not isinstance(task, dict):
                continue
            tname = task.get("name", "<unnamed>")
            entry = task.get("entry")
            if not entry:
                errors.append(f"assets/interface.json: task '{tname}' missing 'entry'")
            elif entry not in all_base_names:
                errors.append(
                    f"assets/interface.json: task '{tname}' entry '{entry}' not found in base pipeline"
                )
            for opt in task.get("option") or []:
                if opt not in defined_options:
                    errors.append(
                        f"assets/interface.json: task '{tname}' references undefined option '{opt}'"
                    )
            _validate_override_map(
                f"task '{tname}'", task.get("pipeline_override"), all_base_names, errors
            )

        for opt_name, opt_def in (interface.get("option") or {}).items():
            if not isinstance(opt_def, dict):
                continue
            for case in opt_def.get("cases") or []:
                if not isinstance(case, dict):
                    continue
                cname = case.get("name", "<unnamed>")
                _validate_override_map(
                    f"option '{opt_name}.{cname}'",
                    case.get("pipeline_override"),
                    all_base_names,
                    errors,
                )

    if errors:
        print("Validation FAILED:")
        for err in errors:
            print(f"  - {err}")
        return 1

    print(
        f"Validation PASSED: interface.json (PI V2), "
        f"{len(base_nodes)} base nodes, {len(bili_nodes)} bilibili override nodes."
    )
    return 0


if __name__ == "__main__":
    sys.exit(main())
