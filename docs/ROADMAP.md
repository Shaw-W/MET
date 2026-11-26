# MET（Maa Eternal Tree）开发进度计划

> 本文档跟踪 MET 从初始化到正式发布的各阶段任务与验收标准。每完成一项将在此更新勾选状态。

---

## 阶段总览

| 阶段 | 名称 | 状态 | 核心交付物 |
| --- | --- | --- | --- |
| Phase 0 | 规范确认与项目初始化 | 已完成 | `README.md`、`docs/SPEC.md`、`docs/ROADMAP.md`、需求决策确认 |
| Phase 1 | 脚手架与自动打包流水线 | 进行中 | PI V2 骨架、`tools/configure.py`、`tools/install.py`、`tools/validate.py`、GitHub Actions（Windows MXU 自动打包） |
| Phase 2 | 基础框架与每日日常 | 待开始 | 启动游戏、通用异常恢复、每日物资/签到/任务/行会/扫荡 |
| Phase 3 | 核心战斗与活动周回 | 待开始 | 多人战斗、神眠之地、紧那罗击退战、神狱桃园乡、探索活动 |
| Phase 4 | 多服适配与交互优化 | 待开始 | B服资源差异覆盖、PI V2 选项细化、异常恢复加固 |
| Phase 5 | 文档完善与正式发布 | 待开始 | 用户手册、常见问题、完整回归测试、`v1.0.0` Release |

---

## 已确认的核心决策

- **打包平台**：仅 Windows (`x86_64`)
- **服区与控制器**：官服 + B服；ADB + Win32
- **逻辑架构**：纯 Pipeline JSON（不引入 Python Agent，不实现已移除的自动炼金）
- **开发路线**：先参考 MAATree 现有流程按 PI V2 与模块化规范重构
- **GitHub 仓库**：[Shaw-W/MET](https://github.com/Shaw-W/MET)（已写入 `interface.json`）

---

## 详细任务清单

### Phase 0：规范确认与项目初始化
- [x] 输出架构规范与开发限制（[docs/SPEC.md](SPEC.md)）
- [x] 输出开发进度计划（[docs/ROADMAP.md](ROADMAP.md)）
- [x] 创建项目入口说明（[README.md](../README.md)）
- [x] 与项目负责人确认待决策事项（目标平台、服区、控制器、纯 Pipeline 方案、开发方式）
- [x] 创建并关联 GitHub 仓库（已将 `https://github.com/Shaw-W/MET` 写入 `assets/interface.json`）

### Phase 1：脚手架与自动打包流水线
- [x] 创建标准目录结构（`.gitignore`、`assets/`、`tools/`、`.github/workflows/`）
- [x] 编写符合 PI V2 规范的 `assets/interface.json` 基础骨架（官服+B服、ADB+Win32、任务入口占位）
- [x] 编写模块化 Pipeline 骨架文件（`startup.json`、`common.json`、`daily.json`、`battle.json`、`activity.json`）
- [x] 编写 `tools/configure.py`（本地开发一键拉取 Windows 版 MaaFramework、MXU 与 OCR 模型）
- [x] 编写 `tools/install.py`（本地与 CI 共用的 Windows 组装脚本，自动注入版本号）
- [x] 编写 `tools/validate.py`（JSON 语法、PI V2 协议与无 MirrorChyan 校验、Pipeline 节点引用校验）
- [x] 编写 `.github/workflows/ci.yml`（PR / Push 自动静态校验）
- [x] 编写 `.github/workflows/release.yml`（无 MirrorChyan，Windows `x86_64` 自动拉取 MaaFW + MXU + OCR 并打包发布至 GitHub Releases）
- [ ] **里程碑验收**：推送到 GitHub 后 CI 成功产出 `MET-win-x86_64-*.zip`，解压后双击 `mxu.exe` 可正常打开 MET 界面并识别控制器与任务列表

### Phase 2：核心通用能力与每日日常
- [ ] 通用基础节点（`common.json`）：主页判定、返回主页、弹窗关闭、掉线重连、卡死重启、跨夜检测、自动补体
- [ ] 启动游戏（`startup.json`）：官服启动与进入主界面
- [ ] 每日物资领取
- [ ] 行会签到与捐赠
- [ ] 每日素材关卡扫荡一次
- [ ] 困难与极难主线自动扫荡
- [ ] 每日任务奖励领取
- [ ] **里程碑验收**：日常一键串行执行稳定通过，支持异常弹窗与断线恢复

### Phase 3：核心战斗与活动周回
- [ ] 多人战斗自动周回（含自动建房、自动补体、次数/体力控制）
- [ ] 神眠之地自动重复刷
- [ ] 紧那罗击退战自动重复打
- [ ] 神狱桃园乡自动领取奖励
- [ ] 探索活动自动领取奖励
- [ ] **里程碑验收**：各周回模块在连续运行与自动补体场景下无卡死

### Phase 4：多服适配与交互优化
- [ ] B服资源差异覆盖层（`assets/resource/bilibili/`）
- [ ] 细化 PI V2 `option` 交互配置项（如补体策略、关卡选择、扫荡次数等）与界面提示文案
- [ ] 异常恢复与长时间挂机稳定性加固

### Phase 5：稳定性打磨与正式发布
- [ ] 编写用户文档 `docs/USAGE.md`（安装、ADB/Win32 连接设置、常见问题、日志导出反馈方式）
- [ ] 全流程端到端回归测试
- [ ] 发布 `v0.1.0-beta` 试运行版本并收集反馈
- [ ] 正式发布 `v1.0.0`
