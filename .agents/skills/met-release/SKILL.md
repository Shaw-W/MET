---
name: met-release
description: >-
  Use this skill when the user asks to release a new version of MET, create a version tag,
  verify GitHub Actions Windows x86_64 MXU packaging, or publish a GitHub Release.
---

# MET 发布（Release）标准流程

MET 仅打包 Windows `x86_64`（MaaFramework + MXU + OCR），通过 GitHub Actions 自动发布至 GitHub Releases，**不包含任何 MirrorChyan 步骤**。

## 1. 发布前检查

1. 确认所有待发布改动已保存，且运行静态校验通过：
   ```bash
   uv run --python 3.12 python tools/validate.py
   ```
2. 检查 `assets/interface.json`：
   - `"interface_version": 2`
   - `"github": "https://github.com/Shaw-W/MET"`
   - 不含 `mirrorchyan_*` 或 `custom_title` 字段。
3. 按需更新 `docs/ROADMAP.md` 中的进度勾选状态。

## 2. 版本号规则

- **预发布版本（Pre-release）**：包含 `alpha`、`beta`、`rc` 或 `dev`，例如 `v0.1.0-alpha.1`、`v0.2.0-beta.1`。CI 会自动标记为 GitHub Pre-release。
- **正式版本（Release）**：标准语义化版本，例如 `v1.0.0`。CI 会自动设为 Latest Release，供 MXU 客户端检测更新。

## 3. 合并到 `main` 并打标签推送

> 注意：`main` 分支已开启保护，禁止直接推送。请先将发布准备改动通过分支提交 PR 并合并至 `main`，再基于最新 `main` 打标签推送：

```bash
git checkout main
git pull origin main
git tag <tag>
git push origin <tag>
```

## 4. 验收发布结果

1. 检查 GitHub Actions 中 `CI Validation` 与 `Build & Release (Windows x86_64)` 均变为绿色（`success`）。
2. 打开 `https://github.com/Shaw-W/MET/releases/tag/<tag>`，确认：
   - 产物 `MET-win-x86_64-<tag>.zip` 已自动上传；
   - Pre-release 标记与自动生成的 Release Notes 正确；
   - 解压后目录包含 `mxu.exe`、`maafw/`、`interface.json`（版本号已注入）与 `resource/`。
