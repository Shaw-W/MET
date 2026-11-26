---
name: maa-pipeline-dev
description: >-
  Use this skill when designing, implementing, refactoring (from MAATree), or debugging
  MaaFramework Pipeline JSON nodes, 720p image templates, ROI coordinates, or PI V2
  interface.json tasks/options for MET, including live debugging via MaaMCP.
---

# MET MaaFramework Pipeline 开发与调试指南

## 1. 模块划分与命名约定

所有基础节点位于 `assets/resource/base/pipeline/`，按功能域拆分为 5 个文件：

| 文件 | 节点前缀 | 职责范围 |
| --- | --- | --- |
| `common.json` | `Common_` | 主页判定、返回主页、弹窗关闭、掉线重连、卡死重启、跨夜检测、自动补体 |
| `startup.json` | `Startup_` | 启动游戏、登录与公告/礼包弹窗清理 |
| `daily.json` | `Daily_` | 每日物资、行会签到与捐赠、素材关卡扫荡、困难/极难主线扫荡、每日任务领奖 |
| `battle.json` | `Battle_` | 多人战斗（建房/周回）、神眠之地、紧那罗击退战 |
| `activity.json` | `Activity_` | 神狱桃园乡领奖、探索活动领奖 |

- **B服差异**：仅将与官服不同的节点或图片放入 `assets/resource/bilibili/pipeline/` 或 `assets/resource/bilibili/image/`，禁止整份复制 `base`。
- **任务入口**：只有顶层用户可见任务写入 `assets/interface.json` 的 `task[]`；内部子步骤不直接暴露。

## 2. Pipeline 编写最佳实践

1. **优先 OCR，其次 TemplateMatch**：
   - 文字稳定的按钮/标题优先用 `recognition: "OCR"` + 精确 `roi: [x, y, w, h]` + `expected` 正则或文字数组。
   - 图标/无文字控件使用 `recognition: "TemplateMatch"` + 720p 无损 PNG 模板（存放于 `assets/resource/base/image/<module>/`）+ 明确 `threshold`（通常 `0.7`~`0.85`）。
2. **统一 720p 基准**：
   - 所有 `roi` 坐标与模板截图必须基于 **1280×720（短边 720）**，不得混用 1080p/2K 坐标。
3. **弹窗与异常恢复**：
   - 善用 `[JumpBack]` 前缀处理偶发弹窗、体力不足提示或掉线重连，确保处理完后回到原状态机继续执行。
   - 合理设置 `post_delay`（如 500~1000ms）与 `timeout`，避免动画过渡期误判。
4. **纯 Pipeline 原则**：
   - 严禁引入 Python `MaaAgent` 或自定义外部脚本；所有逻辑均通过 MaaFramework 原生识别/动作/状态流转表达。

## 3. 使用 MaaMCP 实时采集与验证

当用户启动了《悠久之树》（ADB 模拟器或 Win32 窗口）且本机已启用 `MaaMCP` 时：

1. **扫描与连接**：
   - ADB：`find_adb_device_list` -> `connect_adb_device`（保持默认 `target_short_side=720`）
   - Win32：`find_window_list`（筛选 `悠久之树`）-> `connect_window`（保持默认 `target_short_side=720`）
2. **优先 OCR 探测**：
   - 调用 `ocr` 获取界面文本与边界框 `[x, y, w, h]`，直接换算为精确 `roi`（可适当向外扩 10~20px 留余量）。
3. **按需截屏与存模板**：
   - 调用 `screencap` 查看画面；如需制作 `TemplateMatch` 模板，裁剪后以无损 PNG 保存至 `assets/resource/base/image/<module>/<name>.png`。
4. **验证**：
   - 编辑完 JSON 后必须执行静态校验：
     ```bash
     uv run --python 3.12 python tools/validate.py
     ```
   - 同步勾选 `docs/ROADMAP.md` 中对应任务的进度。
