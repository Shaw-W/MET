# MET（Maa Eternal Tree）开发进度计划

> 本文档跟踪 MET 从初始化到正式发布的各阶段任务与验收标准。每完成一项将在此更新勾选状态。

---

## 阶段总览

| 阶段 | 名称 | 状态 | 核心交付物 |
| --- | --- | --- | --- |
| Phase 0 | 规范确认与项目初始化 | 已完成 | `README.md`、`docs/SPEC.md`、`docs/ROADMAP.md`、需求决策确认 |
| Phase 1 | 脚手架与自动打包流水线 | 已完成 | PI V2 骨架、`tools/configure.py`、`tools/install.py`、`tools/validate.py`、GitHub Actions（Windows MXU 自动打包已验证通过） |
| Phase 2 | 基础框架与每日日常 | 进行中 | 启动游戏（官服+B服）、通用异常恢复、每日物资/行会/扫荡/任务/邮件（代码已重构，待实机调优验收） |
| Phase 3 | 核心战斗与活动周回 | 进行中 | 多人战斗（建房/蹭车）、神眠之地、神狱桃园乡、探索活动（代码已重构，紧那罗击退战暂缓） |
| Phase 4 | 多服适配与交互优化 | 已完成 | B服包名差异覆盖、PI V2 选项细化（素材次数/主线双难度/神眠首领/多人模式/桃园属性）、校验器强化 |
| Phase 5 | 文档完善与正式发布 | 进行中 | 用户手册 `docs/USAGE.md` 已完成，待实机端到端回归测试后发布 `v0.1.0` |

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
- [x] **里程碑验收**：推送到 GitHub 后 CI 成功产出 `MET-win-x86_64-*.zip`（已完成首次打包绿灯与分支保护）

### Phase 2：核心通用能力与每日日常
- [x] 通用基础节点（`common.json`）：主页判定、返回主页、弹窗关闭、掉线重连、卡死重启、跨夜检测、通用扫荡与自动补体
- [x] 启动游戏（`startup.json` + `bilibili/pipeline/startup.json`）：官服与 B服启动、包名差异覆盖与弹窗清理
- [x] 每日物资领取（`Daily_ClaimSupply`）
- [x] 行会签到与捐赠（`Daily_GuildSignIn`）
- [x] 每日素材关卡扫荡一次（`Daily_MaterialSweep`）
- [x] 困难与极难主线自动扫荡（`Daily_MainStorySweep` + `MainStoryDifficulty` 选项）
- [x] 每日任务奖励领取（`Daily_ClaimQuests`）与一键领取邮件（`Daily_ClaimMail`）
- [ ] **里程碑验收**：连接真实游戏窗口/模拟器（通过 MaaMCP 或 MXU）逐项跑通并微调 720p ROI 与偶发弹窗模板

### Phase 3：核心战斗与活动周回
- [x] 多人战斗自动周回（`Battle_Multi_Start`，支持自动建房 / 蹭车模式 `MultiBattleMode` 与自动补体）
- [x] 神眠之地自动重复刷（`Battle_GodSleep_Start`，含材料不足检测与补体联动）
- [ ] 紧那罗击退战自动重复打（**按需求暂缓实现**，保留占位入口）
- [x] 神狱桃园乡扫荡与领奖（`Activity_PeachGarden_Claim`，支持朱雀/玄武/青龙/白虎属性选择）
- [x] 探索活动自动领取奖励（`Activity_Explore_Claim`，覆盖探求者之书与器灵讨伐书日常/周常/奖励）
- [ ] **里程碑验收**：各周回模块在真实游戏环境下连续运行与自动补体场景验证无卡死

### Phase 4：多服适配与交互优化
- [x] B服资源差异覆盖层（`assets/resource/bilibili/pipeline/startup.json` 增量覆盖包名与入口 Activity）
- [x] 补齐 `MAATree` 全部剩余待迁移模块（每日在线奖励、自动助战任务、自动推关、等待被邀请、限时活动副本/商店/任务、天使诞生与仪式提交）
- [x] 细化 PI V2 `option` 交互配置项（补体门控、主线困难/极难/双难度、素材 1次/最大、多人战建房/蹭车、神眠首领、神狱桃园乡属性）
- [x] 异常恢复与长时间挂机稳定性加固（战斗超时收敛、自动战斗检测、蹭车超时退房、神眠 SAP 兑换、`tools/validate.py` 校验 `pipeline_override`）

### Phase 5：稳定性打磨与正式发布
- [x] 编写用户文档 [docs/USAGE.md](USAGE.md)（安装、ADB/Win32 连接设置、任务选项、常见问题与日志反馈指引）
- [ ] 实机端到端回归测试（ADB 模拟器 + Win32 PC 端逐项微调 ROI）
- [ ] 发布 `v0.1.0` 首个完整功能版本并收集反馈
- [ ] 正式发布 `v1.0.0`
