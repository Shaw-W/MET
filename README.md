<!-- markdownlint-disable MD033 MD041 -->

<div align="center">

<img alt="MET Logo" src="./assets/logo.png" width="180" height="180" />

# MET (Maa Eternal Tree)

<p>
  <img alt="Platform" src="https://img.shields.io/badge/platform-Windows%20x86__64-0078D6?logo=windows&logoColor=white" />
  <img alt="Engine" src="https://img.shields.io/badge/MaaFramework-%3E%3D%20v5.12.1-2ea44f" />
  <img alt="UI" src="https://img.shields.io/badge/GUI-MXU%20(PI%20V2)-blueviolet" />
  <img alt="Architecture" src="https://img.shields.io/badge/Pipeline-Pure%20JSON-454545" />
</p>
<p>
  <img alt="License" src="https://img.shields.io/github/license/Shaw-W/MET" />
  <img alt="Commit Activity" src="https://img.shields.io/github/commit-activity/m/Shaw-W/MET?color=%23ff69b4" />
  <img alt="Release" src="https://img.shields.io/github/v/release/Shaw-W/MET?include_prereleases" />
  <img alt="Stars" src="https://img.shields.io/github/stars/Shaw-W/MET?style=social" />
</p>

面向《悠久之树》的新一代轻量化自动化小助手

基于 **[MaaFramework](https://github.com/MaaXYZ/MaaFramework)** 纯 Pipeline JSON 架构与 **[MXU](https://github.com/MaaXYZ/MXU)** 前端构建，零外部运行时依赖，解压即用，解放双手！

[架构规范](docs/SPEC.md) · [开发进度](docs/ROADMAP.md) · [AI 开发约定](AGENTS.md) · [下载发布包](https://github.com/Shaw-W/MET/releases)

</div>

---

## 快速开始

本仓库仅保留**配置、Pipeline 源码与模板素材**，运行时二进制（`mxu.exe`、`maafw/` 与 OCR 模型）均由 GitHub Actions 自动拉取并打包。

- **普通用户**：前往 [Releases](https://github.com/Shaw-W/MET/releases) 下载最新 `MET-win-x86_64-*.zip`，解压后直接双击 `mxu.exe` 即可运行，无需安装 Python 或其他运行环境。
- **支持环境**：Windows (`x86_64`)，支持 **官服 + B服**，支持 **ADB（安卓模拟器，推荐 `1280×720`）** 与 **Win32（PC 桌面端）**。

---

## 计划支持的功能

- **日常自动化**：启动游戏、每日物资、行会签到与捐赠、素材关卡扫荡、困难与极难主线扫荡、每日任务领奖
- **战斗与活动周回**：多人战斗、神眠之地、紧那罗击退战、神狱桃园乡领奖、探索活动领奖
- **稳定性保障**：自动补体、自动建房、卡死重启、掉线重连、自动跨夜

详细任务拆分与当前里程碑请查看 [开发进度计划](docs/ROADMAP.md)。

---

## 本地开发与校验

```bash
# 1. 静态校验 PI V2 配置与 Pipeline 节点引用
uv run --python 3.12 python tools/validate.py

# 2. 本地拉取 OCR 模型及 Windows 版 MaaFramework / MXU
uv run --python 3.12 python tools/configure.py --download-runtime

# 3. 组装本地可运行目录（输出至 install/）
uv run --python 3.12 python tools/install.py v0.0.0-dev
```

---

## 参与开发与生态工具链

### 核心关联项目

- 自动化引擎：[MaaFramework](https://github.com/MaaXYZ/MaaFramework)
- 图形界面前端：[MXU (MaaFramework Next UI)](https://github.com/MaaXYZ/MXU)
- 公共 OCR 模型与资源：[MaaCommonAssets](https://github.com/MaaXYZ/MaaCommonAssets)
- 实时设备/窗口 MCP 调试：[MaaMCP](https://github.com/MaaXYZ/MaaMCP)

### 推荐开发工具链

- [MaaPipelineEditor](https://github.com/kqcoxn/MaaPipelineEditor)：可视化构建、审阅、编辑与调试 MaaFramework Pipeline
- [Everything-Maa](https://github.com/KhazixW2/Everything-Maa)：面向 MaaFramework 的技能、MCP 集成与 AI 辅助开发工具集
- [maa-tools](https://www.npmjs.com/package/@nekosu/maa-tools)：`pnpm check` / `pnpm format` 背后的格式化与 Schema 校验工具

### 贡献指引

- 所有架构约束与红线详见 [docs/SPEC.md](docs/SPEC.md) 与 [AGENTS.md](AGENTS.md)。
- Pipeline 节点按模块存放于 `assets/resource/base/pipeline/`，B服差异仅增量放置于 `assets/resource/bilibili/`。
- 提交 PR 前请确保 `tools/validate.py` 校验通过，并同步更新 [docs/ROADMAP.md](docs/ROADMAP.md)。

---

## 鸣谢

- 核心引擎与生态：[MaaFramework](https://github.com/MaaXYZ/MaaFramework)、[MXU](https://github.com/MaaXYZ/MXU)、[MaaCommonAssets](https://github.com/MaaXYZ/MaaCommonAssets)
- OCR 识别模型：[PaddleOCR](https://github.com/PaddlePaddle/PaddleOCR)
- 功能流程参考：[MAATree](https://github.com/caicai00001/MAATree)

感谢每一位参与测试、反馈与贡献代码的朋友！

---

## 声明

- 本项目免费开源，仅供学习与技术交流使用。
- 本软件与游戏官方无关，仅通过图像识别与模拟输入与游戏用户界面交互，不读取或修改任何游戏内存、文件与网络协议；请自行评估使用风险。
- 本软件严禁用于商业代练或收费分发，由此产生的任何问题及后果均与本项目及开发者无关。
- 本项目不拥有原游戏美术资源版权，相关素材版权归原游戏厂商所有。

---

<div align="center">

如果 MET 对你有帮助，欢迎点一颗右上角的 **Star** 支持一下！

</div>
