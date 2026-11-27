<!-- markdownlint-disable MD033 MD041 -->

<div align="center">

<img alt="MET Logo" src="./assets/logo.png" width="180" height="180" />

# MET (Maa Eternal Tree)

<p>
  <img alt="Platform" src="https://img.shields.io/badge/platform-Windows%20x86__64-0078D6?logo=windows&logoColor=white" />
  <img alt="Engine" src="https://img.shields.io/badge/MaaFramework-%3E%3D%20v5.12.1-2ea44f" />
  <img alt="UI" src="https://img.shields.io/badge/GUI-MXU%20(PI%20V2)-blueviolet" />
  <img alt="Pipeline" src="https://img.shields.io/badge/Pipeline-JSON-454545" />
  <a href="./agent"><img alt="Go" src="https://img.shields.io/badge/Go-007d9c?logo=go&logoColor=white" /></a>
</p>
<p>
  <a href="https://github.com/Shaw-W/MET/actions/workflows/ci.yml"><img alt="CI Validation" src="https://github.com/Shaw-W/MET/actions/workflows/ci.yml/badge.svg" /></a>
  <a href="https://github.com/Shaw-W/MET/actions/workflows/release.yml"><img alt="Build & Release" src="https://github.com/Shaw-W/MET/actions/workflows/release.yml/badge.svg" /></a>
  <a href="./LICENSE"><img alt="License" src="https://img.shields.io/badge/license-AGPL--3.0-blue" /></a>
  <a href="https://github.com/Shaw-W/MET/stargazers"><img alt="Stars" src="https://img.shields.io/github/stars/Shaw-W/MET?style=social" /></a>
</p>

面向《悠久之树》的新一代轻量化自动化小助手

基于 **[MaaFramework](https://github.com/MaaXYZ/MaaFramework)** Pipeline JSON + 轻量 **Go** 服务与 **[MXU](https://github.com/MaaXYZ/MXU)** 前端构建，零外部运行时依赖，解压即用，解放双手！

[使用手册](docs/USAGE.md) · [架构规范](docs/SPEC.md) · [开发进度](docs/ROADMAP.md) · [AI 开发约定](AGENTS.md) · [下载发布包](https://github.com/Shaw-W/MET/releases)

</div>

---

## 快速开始

本仓库仅保留**配置、Pipeline 源码、Go 辅助服务源码与模板素材**，运行时二进制（`MET.exe`、`agent/met-service.exe`、`maafw/` 与 OCR 模型）均由 GitHub Actions 自动编译拉取并打包。

- **普通用户**：前往 [Releases](https://github.com/Shaw-W/MET/releases) 下载最新 `MET-win-x86_64-*.zip`，解压后直接双击 `MET.exe` 即可运行，无需安装 Python 或其他运行环境。
- **支持环境**：Windows (`x86_64`)，支持 **官服 + B服**，支持 **ADB（安卓模拟器，推荐 `1280×720`）** 与 **Win32（PC 桌面端）**。

---

## ✨ 功能一览

- 🚀 **启动游戏**：PC 端自动检测并拉起 `eternaltree.exe`（已在运行则秒过），支持 ADB 官服 / B服自动启动，定时轮询并自动清理登录与礼包弹窗，稳稳进主界面！
- 🔋 **体力补给**：前置进入「活动 → 福利放送」同页连续领取每日物资（12点/18点）与在线奖励，支持独立开关，无奖自动提前结束。
- 💎 **月卡扫荡**：自动进入主页「月卡 → 特许关卡」一键扫荡全部剩余特许关卡，无剩余次数自动跳过。
- 🏰 **行会日常**：自动从菜单抽屉进入行会完成每日签到与物资捐赠，支持分别开关、真正的每日捐赠上限（1～3 次）及已完成自动跳过。
- ⚔️ **联机自动战斗**：一键进入「联机 → 自动战斗」，支持按 MET 选项自动同步 6 项战斗开关与精力药水勾选，内置防卡死守护（画面静止超 5 分钟自动强杀重启 PC 客户端并重连续刷，异常掉回联机大厅自动续战）。
- 🎁 **领取奖励**：收尾统一清红点——一键领取邮箱附件、任务（日常含活跃宝箱 / 主线 / 成就三分类）、探求者之书 & 器灵讨伐书（先领任务经验再领等级奖励）以及探求者手记，全部支持逐项开关 ✅
- 💻 **多端与多服适配**：支持 **Win32 PC 端**（原生兼容 1080p / 2K 16:9 缩放）与 **ADB 安卓模拟器**，支持 **官服 + B服**，内置「核心日常（推荐）」一键预设。
- 🚧 **正在维护**：素材 / 主线扫荡、神眠之地、神狱桃源乡与限时活动等模块正在按 4.0 界面陆续适配中，详见 [开发进度计划](docs/ROADMAP.md)。

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

- 本项目基于 [GNU AGPL-3.0](LICENSE) 协议免费开源，仅供学习与技术交流使用。
- 本软件与游戏官方无关，仅通过图像识别与模拟输入与游戏用户界面交互，不读取或修改任何游戏内存、文件与网络协议；请自行评估使用风险。
- 严禁将本项目用于商业代练或闭源收费分发；任何基于本项目的二次开发或网络/云端部署均须严格遵守 AGPL-3.0 协议并开源对应源码，由此产生的任何问题及后果均与本项目及开发者无关。
- 本项目不拥有原游戏美术资源版权，相关素材版权归原游戏厂商所有。

---

<div align="center">

如果 MET 对你有帮助，欢迎点一颗右上角的 **Star** 支持一下！

</div>
