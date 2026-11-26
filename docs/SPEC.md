# MET（Maa Eternal Tree）架构规范与开发限制

> 本文档定义 **MET（Maa Eternal Tree）** 的技术架构、目录规范、协议标准、CI/CD 自动打包要求及开发红线。所有后续代码、资源文件与流水线配置均须严格遵守本规范。

---

## 1. 项目定位与核心技术栈

MET 是面向《悠久之树》的新一代自动化助手，参考 [MAATree](https://github.com/caicai00001/MAATree) 的功能设计，基于 [MaaXYZ](https://github.com/MaaXYZ) 生态最新标准重构：

| 层级 | 选型 | 版本/要求 | 说明 |
| --- | --- | --- | --- |
| 核心引擎 | [MaaFramework](https://github.com/MaaXYZ/MaaFramework) | `>= v5.12.1` | 图像识别、OCR、模拟控制、Pipeline 调度引擎 |
| 接口协议 | **Project Interface V2 (PI V2)** | `"interface_version": 2` | 前后端解耦协议，严格遵循 `interface.schema.json` |
| 前端 GUI | [MXU (MaaFramework Next UI)](https://github.com/MaaXYZ/MXU) | Latest Release | Tauri + React + TypeScript 通用客户端（仅打包 Windows `x86_64`） |
| 逻辑实现 | **纯 Pipeline JSON** | 零外部运行时依赖 | 不引入 Python `MaaAgent`（自动炼金功能已移除），保持产物轻量便携 |
| 目标平台与服区 | Windows (`x86_64`) / 官服 + B服 | ADB + Win32 | 支持安卓模拟器（ADB）与 PC 桌面端（Win32） |
| 自动化打包 | GitHub Actions | `win-x86_64` 自动组装 | 仅发布至 GitHub Actions Artifacts / GitHub Releases，不含 MirrorChyan |

---

## 2. 目录结构规范

### 2.1 开发态仓库结构（源码目录）

```text
MET/
├── .github/
│   └── workflows/
│       ├── ci.yml                 # PR / Push 静态校验（JSON 语法、PI V2 字段、Pipeline 引用检查）
│       └── release.yml            # Windows 自动打包流水线（拉取 MaaFW + MXU + OCR，发布 GitHub Release）
├── assets/
│   ├── interface.json             # PI V2 协议主配置文件（开发态）
│   └── resource/
│       ├── base/                  # 基础通用资源（官服默认，含 pipeline/、image/、model/ocr/）
│       │   ├── pipeline/          # 按功能模块拆分的 Pipeline JSON
│       │   ├── image/             # 模板匹配截图（PNG）
│       │   └── model/ocr/         # OCR 模型（CI 自动拉取或本地 configure.py 拉取，不提交大二进制）
│       └── bilibili/              # B服差异覆盖资源（仅放与 base 不同的图片/节点）
├── docs/
│   ├── SPEC.md                    # 本文档：架构规范与开发限制
│   ├── ROADMAP.md                 # 开发进度计划与里程碑
│   └── USAGE.md                   # （后续补充）用户使用说明与常见问题
├── tools/
│   ├── configure.py               # 本地开发环境初始化（拉取 OCR 模型 / MaaFW / MXU）
│   ├── install.py                 # 构建组装脚本（输出符合 MXU 规范的 Windows 运行目录）
│   └── validate.py                # 静态校验脚本（供本地与 CI 共用）
├── .gitignore
└── README.md
```

### 2.2 发布态运行结构（MXU Windows 打包产物）

打包脚本 `tools/install.py` 与 CI 流水线的输出必须严格符合 MXU 的目录布局要求：

```text
MET-win-x86_64-<version>/
├── mxu.exe                        # MXU Windows x86_64 可执行文件
├── maafw/                         # MaaFramework Release 中的 bin/ 全部内容
│   ├── MaaFramework.dll
│   ├── MaaToolkit.dll
│   └── ...
├── interface.json                 # 已注入版本号的 PI V2 配置文件
└── resource/
    ├── base/
    └── bilibili/
```

---

## 3. 协议与资源规范

### 3.1 PI V2 (`assets/interface.json`) 规范

1. **必须声明协议版本**：顶层必须包含 `"interface_version": 2`。
2. **严格标准 JSON**：不允许包含 `//` 或 `/* */` 注释及尾随逗号，必须通过 MaaFramework 官方 `interface.schema.json` 校验。
3. **禁止遗留与无关字段**：
   - **严禁包含** `mirrorchyan_rid`、`mirrorchyan_multiplatform` 等 MirrorChyan 字段。
   - **严禁包含** MFAAvalonia / PI V1 专属字段（如 `custom_title`），窗口标题与显示名统一使用 PI V2 的 `name`、`label`、`title` 字段。
4. **版本号管理**：开发态 `version` 占位符为 `"dev"` 或 `"0.0.0"`，由 `tools/install.py` 在打包时根据 Git Tag / Commit 自动注入。
5. **控制器配置**：
   - ADB 控制器默认配置 `"display_short_side": 720`（16:9 基准即 `1280x720`），保证跨设备坐标与模板匹配一致性。
   - Win32 控制器（如支持桌面端/模拟器窗口）按需配置 `window_regex`、`class_regex` 及截图/输入方式。
6. **资源分层覆盖**：
   - 官服加载 `["{PROJECT_DIR}/resource/base"]`。
   - B服（如支持）加载 `["{PROJECT_DIR}/resource/base", "{PROJECT_DIR}/resource/bilibili"]`，增量覆盖差异项，禁止整份复制 `base`。

### 3.2 Pipeline 与图像资源规范

1. **模块化拆分**：`assets/resource/base/pipeline/` 下按功能域拆分 JSON 文件，严禁将全部节点堆砌在单个超大文件中。建议划分：
   - `startup.json`（启动游戏、登录、公告/礼包弹窗关闭）
   - `common.json`（返回主页、掉线重连、卡死重启、跨夜检测、自动补体）
   - `daily.json`（每日物资、行会签到捐赠、每日任务、主线/素材扫荡）
   - `battle.json`（多人战斗、自动建房、神眠之地、紧那罗击退战）
   - `activity.json`（神狱桃园乡、探索活动）
2. **节点命名规范**：
   - 统一采用 `模块前缀_语义名称`（如 `Daily_ClaimSupply`、`Common_BackHome`、`Battle_Multi_CreateRoom`），避免跨文件节点名冲突。
   - 仅作为子步骤/内部跳转的节点不直接暴露在 `interface.json` 的 `task` 中。
3. **图像与 ROI 规范**：
   - 所有模板截图必须采集自 **1280×720（短边 720p）** 标准分辨率，格式为无损 PNG。
   - 截图按模块存放在 `image/<module>/` 子目录中，文件名清晰可读，禁止无意义的随机命名。
   - 识别节点优先限定精确的 `roi` 区域，减少全屏匹配误触发并提升识别速度。
   - 优先使用 `OCR` 或 `TemplateMatch` + 合理阈值（`threshold`），对动态背景区域避免盲目放大 `roi`。

### 3.3 纯 Pipeline 原则（不引入 Agent）

1. **零 Python Agent**：本项目不引入 `MaaAgent` 自定义代码，不实现已移除的“自动炼金”功能，所有任务均通过 MaaFramework 原生 Pipeline JSON 实现，确保产物开箱即用、零额外运行时依赖。
2. **流程参考策略**：开发时可先参考 MAATree 的现有业务流程与状态流转设计，按本规范的模块化命名与 PI V2 协议进行重构，并结合现版本《悠久之树》界面更新素材与节点。

---

## 4. CI/CD 自动打包规范

1. **移除 MirrorChyan 流水线**：
   - 仓库中**不创建、不保留**任何 `mirrorchyan_release.yml`、`MirrorChyan/uploading-action` 步骤或相关 Secrets。
2. **自动拉取上游组件**：
   - 在 GitHub Actions 中支持通过环境变量固定或自动拉取最新版本的：
     - `MaaXYZ/MaaFramework`（提取 `bin/` 至产物 `maafw/`）
     - `MaaXYZ/MXU`（提取对应平台可执行文件/App 包至产物根目录）
     - OCR 模型资源（`MaaXYZ/MaaCommonAssets` 对应中文模型，不把大体积模型直接提交进 Git 仓库）
3. **触发机制**：
   - **PR / Push 到主分支**：执行静态校验（JSON 语法、PI V2 规范校验、Pipeline 节点引用完整性检查）+ 构建 Windows `x86_64` 测试包并上传为 GitHub Actions Artifact。
   - **Push Tag (`v*`) / 手动 `workflow_dispatch`**：执行 Windows `x86_64` 打包流程，生成 `MET-win-x86_64-<version>.zip` 并自动发布至 GitHub Releases（含自动区分正式版与 Pre-release）。
4. **版本注入**：
   - `tools/install.py` 统一负责本地与 CI 的目录组装、版本号写入 `interface.json`、清理无关临时文件，确保本地构建与 CI 构建行为 100% 一致。

---

## 5. 开发限制与红线（Constraints）

| 编号 | 限制项 | 具体要求 |
| --- | --- | --- |
| C-01 | **合规与非侵入原则** | 严禁读取、修改、注入游戏内存/进程/本地存档或抓包改包；仅允许通过 MaaFramework 的图像识别与模拟输入与游戏 UI 交互。 |
| C-02 | **零 MirrorChyan 耦合** | 不引入 MirrorChyan 工作流、不写入 `mirrorchyan_*` 字段、不在文档中引导配置 MirrorChyan。 |
| C-03 | **纯 MXU + PI V2 架构** | 不引入 MFAAvalonia、MaaPiCli 专用变通逻辑或 PI V1 废弃语法；所有 UI 展示与配置均通过标准 PI V2 声明。 |
| C-04 | **无硬编码本地路径** | 禁止在 Pipeline、脚本或配置中写死开发者本机绝对路径（如 `D:\...`、`C:\Users\...`）；统一使用 `{PROJECT_DIR}` 或相对路径。 |
| C-05 | **仓库体积控制** | 禁止将 `maafw/` 动态库、`mxu.exe`、OCR 模型大文件、日志文件、本地调试截图提交到 Git（须在 `.gitignore` 中严格排除）。 |
| C-06 | **分辨率基准统一** | 所有模板图与 ROI 坐标均以短边 `720`（1280×720）为基准，不得混入 1080p/2K 原图导致识别失效。 |
| C-07 | **变更可验证** | 每次新增或修改 Pipeline / `interface.json` 后，必须通过 JSON 格式与 Schema 校验，且不破坏现有任务的节点引用。 |
