# AGENTS.md — MET (Maa Eternal Tree) AI 开发规范

本文件为所有参与 **MET（Maa Eternal Tree）** 开发的 AI 编程助手提供项目级上下文、架构约束与标准工作流。

---

## 1. 项目概览与技术栈

- **目标游戏**：《悠久之树》（官服 + B服）
- **核心引擎**：[MaaFramework](https://github.com/MaaXYZ/MaaFramework) (`>= v5.12.1`)
- **接口协议**：**Project Interface V2 (PI V2)**（`"interface_version": 2`）
- **前端界面**：[MXU](https://github.com/MaaXYZ/MXU)（仅打包 Windows `x86_64`）
- **控制器**：ADB（安卓模拟器）+ Win32（桌面端），统一基准分辨率为 **短边 720（1280×720）**
- **仓库地址**：https://github.com/Shaw-W/MET

---

## 2. 必须遵守的硬性约束

1. **纯 Pipeline JSON 架构**：不引入 Python `MaaAgent` 自定义代码，不实现已移除的“自动炼金”功能，确保最终产物零外部运行时依赖。
2. **零 MirrorChyan 耦合**：严禁添加任何 `mirrorchyan_*` 字段、MirrorChyan 工作流或相关文档。
3. **纯 PI V2 + MXU 协议**：`assets/interface.json` 必须是严格标准 JSON（无注释、无尾随逗号），顶层固定 `"interface_version": 2`，严禁引入 MFAAvalonia / PI V1 废弃字段（如 `custom_title`）。
4. **模块化 Pipeline**：
   - `assets/resource/base/pipeline/` 按模块维护：`common.json`、`startup.json`、`daily.json`、`battle.json`、`activity.json`。
   - 节点命名统一采用 `模块前缀_语义名称`（如 `Daily_ClaimSupply`、`Common_BackHome`），禁止跨文件重名。
   - B服差异仅放入 `assets/resource/bilibili/` 做增量覆盖，禁止复制整份 `base`。
5. **720p 图像与 ROI 规范**：所有模板图（无损 PNG）和 ROI 坐标均以短边 `720`（1280×720）采集和编写；优先使用精确 `roi` 与 `OCR` / `TemplateMatch`。
6. **无硬编码路径 & 不提交二进制产物**：禁止写死本机绝对路径；禁止将 `deps/`、`install/`、`dist/`、`maafw/`、`mxu.exe`、OCR 模型或调试日志提交至 Git。
7. **非侵入式合规**：仅通过 MaaFramework 图像识别与模拟输入与游戏 UI 交互，严禁任何内存读写、注入或抓包改包行为。

---

## 3. 常用命令与校验

修改任何 `assets/interface.json` 或 Pipeline JSON 后，**必须**运行静态校验：

```bash
python tools/validate.py
# 若本机使用 uv：
uv run python tools/validate.py
```

本地组装 Windows MXU 运行目录（先拉取依赖再组装）：

```bash
python tools/configure.py --download-runtime
python tools/install.py v0.0.0-dev
```

---

## 4. 专项技能（`.agents/skills/`）与实时调试

仓库内置了按需加载的 Workspace Skills，执行对应任务时请先读取相应 `SKILL.md`：
- **Pipeline 开发与 MaaMCP 调试**：`.agents/skills/maa-pipeline-dev/SKILL.md`
- **版本打包与 Release 发布**：`.agents/skills/met-release/SKILL.md`

如本机已启用 `MaaMCP` MCP 服务器，请始终保持默认 `target_short_side=720` 连接 ADB 或 Win32（窗口正则 `([Ee]ternal\s*[Tt]ree|悠久之树)`），优先用 `ocr` 获取文字与坐标框，修改节点后必须执行 `tools/validate.py` 并更新 `docs/ROADMAP.md`。

---

## 5. 分支保护与发布（Release）流程

- **`main` 分支保护**：禁止直接向 `main` 推送提交，所有改动须在功能/维护分支（如 `feat/*`、`fix/*`、`chore/*`）提交并推送，通过 Pull Request 经 `CI Validation` (`validate`) 校验后合并。
- 合并到 `main`：自动触发 `CI Validation` 与 `Build & Release (Windows x86_64)`，上传测试包 Artifact。
- 推送 `v*` 标签（如 `v0.1.0-alpha.1` 或 `v1.0.0`）：自动打包 `MET-win-x86_64-<tag>.zip` 并创建 GitHub Release（含 `alpha`/`beta`/`rc`/`dev` 的标签自动标记为 Pre-release）。
