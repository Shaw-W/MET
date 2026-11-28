# MET（Maa Eternal Tree）使用文档导航

欢迎使用 **MET**！使用文档已按主题拆分为独立页面，方便按需查阅：

## 1. 下载与目录结构

1. 前往 [Releases](https://github.com/Shaw-W/MET/releases) 下载最新 `MET-win-x86_64-*.zip`。
2. 完整解压到纯英文目录（如 `D:\MET\`），确保同级目录包含：
   ```text
   MET/
   ├── MET.exe                 # 主程序（MXU 前端）
   ├── agent/met-service.exe   # Go 辅助服务
   ├── maafw/                  # MaaFramework 核心运行时
   ├── interface.json          # PI V2 配置
   ├── logo.png
   └── resource/               # 模型、Pipeline 与素材
   ```
3. 直接双击 `MET.exe` 运行（零外部运行时依赖，无需安装 Python 或 Go）。

---

## 2. 文档目录

- 🔌 **[连接配置教程（PC Win32 / ADB 模拟器）](CONNECT.md)** — 管理员权限设置、窗口分辨率、自动拉起游戏、MuMu/雷电 ADB 调试配置
- 📋 **[功能与任务配置说明](TASKS.md)** — 4.0 已校准任务一览、联机自动战斗与全员重伤等待规则说明
- ❓ **[常见问题与日志反馈（FAQ）](FAQ.md)** — `debug/` 留档目录说明、常见连接/识别问题排查与 Issue 反馈模板
- 🗺️ **[开发进度计划](ROADMAP.md)** — 当前版本里程碑与后续 4.0 玩法适配排期
- 📐 **[架构规范](SPEC.md)** — 面向开发者的架构约束与目录规范
