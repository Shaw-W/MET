//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

type coopMonitorParam struct {
	FreezeTimeoutMs int `json:"freeze_timeout_ms"`
	PollIntervalMs  int `json:"poll_interval_ms"`
	MaxDurationMs   int `json:"max_duration_ms"`
}

type ocrDetailJSON struct {
	All []struct {
		Text  string  `json:"text"`
		Score float64 `json:"score"`
	} `json:"all"`
}

type CoopAutoMonitorAction struct {
	root string
}

func (a *CoopAutoMonitorAction) saveDebugScreenshot(img image.Image, tag string) string {
	if img == nil {
		return ""
	}
	dir := filepath.Join(a.root, "debug", "coop_freeze")
	_ = os.MkdirAll(dir, 0o755)
	name := fmt.Sprintf("%s_%s.png", time.Now().Format("2006-01-02_15-04-05"), tag)
	fullPath := filepath.Join(dir, name)
	f, err := os.Create(fullPath)
	if err != nil {
		a.logf("保存异常截图失败 (%s): %v", fullPath, err)
		return ""
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		a.logf("编码异常截图失败 (%s): %v", fullPath, err)
		return ""
	}
	a.logf("已保存异常现场截图: %s", fullPath)
	return fullPath
}

func (a *CoopAutoMonitorAction) logf(format string, args ...any) {
	logDir := filepath.Join(a.root, "debug")
	_ = os.MkdirAll(logDir, 0o755)
	logPath := filepath.Join(logDir, "met_coop_watchdog.log")
	msg := fmt.Sprintf("[%s] [Agent] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = f.WriteString(msg)
		_ = f.Close()
	}
}

func notifyMXU(ctx *maa.Context, msg string) {
	_, _ = ctx.RunTask("Battle_Coop_Auto_Notify", map[string]any{
		"Battle_Coop_Auto_Notify": map[string]any{
			"focus": map[string]any{
				"Node.Action.Succeeded": msg,
			},
		},
	})
}

func joinOCRText(detailJSON string, minScore float64) string {
	if detailJSON == "" {
		return ""
	}
	var d ocrDetailJSON
	if err := json.Unmarshal([]byte(detailJSON), &d); err != nil {
		return ""
	}
	var parts []string
	for _, item := range d.All {
		t := strings.TrimSpace(item.Text)
		if t != "" && item.Score >= minScore {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func isBattleHudText(text string) bool {
	return strings.Contains(text, "回合") || strings.Contains(text, "时间") || strings.Contains(text, ":")
}

type wipedConfigNode struct {
	Enabled *bool `json:"enabled"`
	Timeout *int  `json:"timeout"`
}

func readWipedWaitConfig(ctx *maa.Context) (enabled bool, wait time.Duration) {
	enabled = true
	wait = 10 * time.Minute
	raw, err := ctx.GetNodeJSON("Battle_Coop_Auto_WipedConfig")
	if err != nil || strings.TrimSpace(raw) == "" {
		return enabled, wait
	}
	var cfg wipedConfigNode
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return enabled, wait
	}
	if cfg.Enabled != nil && !*cfg.Enabled {
		return false, 0
	}
	if cfg.Timeout != nil {
		if *cfg.Timeout < 0 {
			return false, 0
		}
		wait = time.Duration(*cfg.Timeout) * time.Millisecond
	}
	return enabled, wait
}

func sleepInterruptible(tasker *maa.Tasker, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if tasker.Stopping() {
			return false
		}
		rem := time.Until(deadline)
		if rem > 500*time.Millisecond {
			rem = 500 * time.Millisecond
		}
		time.Sleep(rem)
	}
	return !tasker.Stopping()
}

func (a *CoopAutoMonitorAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	param := coopMonitorParam{
		FreezeTimeoutMs: 300000,
		PollIntervalMs:  2000,
		MaxDurationMs:   43200000,
	}
	if strings.TrimSpace(arg.CustomActionParam) != "" {
		_ = json.Unmarshal([]byte(arg.CustomActionParam), &param)
	}
	if param.PollIntervalMs <= 0 {
		param.PollIntervalMs = 2000
	}
	if param.MaxDurationMs <= 0 {
		param.MaxDurationMs = 43200000
	}

	tasker := ctx.GetTasker()
	ctrl := tasker.GetController()
	if ctrl == nil {
		a.logf("错误: 无法获取 Controller 实例")
		return false
	}

	// Always start fresh from 0 on each run; count each completed battle on "战斗结算" rising edge.
	completed := 0
	inSettlement := false
	lobbyStreak := 0
	var wipedSince time.Time
	wipedAutoRetreat, wipedWait := readWipedWaitConfig(ctx)

	startMsg := "⚔️ 联机自动战斗开始看护（按「战斗结算」从 0 实时计数 + 软卡死监控）"
	a.logf("%s [wipedAutoRetreat=%v wipedWait=%v]", startMsg, wipedAutoRetreat, wipedWait)
	notifyMXU(ctx, startMsg)

	startTime := time.Now()
	lastProgressTime := time.Now()
	lastScanLogTime := time.Time{}
	lastHudText := ""
	hudVisible := false
	pollInterval := time.Duration(param.PollIntervalMs) * time.Millisecond
	freezeTimeout := time.Duration(param.FreezeTimeoutMs) * time.Millisecond
	maxDuration := time.Duration(param.MaxDurationMs) * time.Millisecond

	for {
		if tasker.Stopping() {
			a.logf("收到用户停止指令，退出联机看护（本次已完成 %d 场）。", completed)
			return true
		}
		if time.Since(startTime) >= maxDuration {
			msg := fmt.Sprintf("⏱️ 联机自动战斗已达到最大看护时长，本次共完成 %d 场", completed)
			a.logf("%s", msg)
			notifyMXU(ctx, msg)
			_ = ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{{Name: "Common_Stop"}})
			return true
		}

		job, err := ctrl.PostScreencap()
		if err != nil || job == nil {
			a.logf("截屏请求失败: %v", err)
			if !sleepInterruptible(tasker, pollInterval) {
				return true
			}
			continue
		}
		job.Wait()
		img, err := ctrl.CacheImage()
		if err != nil || img == nil {
			a.logf("获取截屏缓存失败: %v", err)
			if !sleepInterruptible(tasker, pollInterval) {
				return true
			}
			continue
		}

		// 1. Check normal finish dialog.
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_FinishedDialog", img); err == nil && rec != nil && rec.Hit {
			msg := fmt.Sprintf("✅ 联机自动战斗已结束，本次共完成 %d 场", completed)
			a.logf("%s", msg)
			notifyMXU(ctx, msg)
			_ = ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{{Name: "Battle_Coop_Auto_FinishedDialog"}})
			return true
		}

		// 2. Check dropped back to coop lobby (require 3 consecutive hits ~6s to ignore brief 1-2s inter-battle lobby flashes).
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_FinishedLobby", img); err == nil && rec != nil && rec.Hit {
			lobbyStreak++
			if lobbyStreak >= 3 {
				msg := fmt.Sprintf("🔄 检测到持续停留在联机大厅（本次已完成 %d 场），准备沿用当前配置续刷", completed)
				a.logf("%s", msg)
				notifyMXU(ctx, msg)
				_ = ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{{Name: "Battle_Coop_Auto_FinishedLobby"}})
				return true
			}
		} else {
			lobbyStreak = 0
		}

		// 3. Check network disconnect.
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_NetworkDisconnect", img); err == nil && rec != nil && rec.Hit {
			a.saveDebugScreenshot(img, "network_disconnect")
			msg := fmt.Sprintf("⚠️ 检测到网络连接断开（本次已完成 %d 场），正在尝试重连续刷...", completed)
			a.logf("%s", msg)
			notifyMXU(ctx, msg)
			_ = ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{{Name: "Battle_Coop_Auto_NetworkDisconnect"}})
			return true
		}

		// 3.5 Check "全员重伤" (all party wiped):
		// Wait for teammates to finish the boss (default 10 min, configurable via CoopWipedWaitTimeout).
		// Only after waiting wipedWait do we select right card "无力战斗，全员撤退" and confirm (never spend revive items).
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_AllWiped", img); err == nil && rec != nil && rec.Hit {
			lastProgressTime = time.Now()
			if wipedSince.IsZero() {
				wipedSince = time.Now()
				a.saveDebugScreenshot(img, "all_wiped")
				if !wipedAutoRetreat {
					msg := fmt.Sprintf("⚠️ 检测到「全员重伤」（已设为不自动撤退，等待队友或倒计时结束，本次已完成 %d 场）", completed)
					a.logf("%s", msg)
					notifyMXU(ctx, msg)
				} else if wipedWait > 0 {
					msg := fmt.Sprintf("⚠️ 检测到「全员重伤」，等待队友输出 %.0f 分钟后再撤退（本次已完成 %d 场）", wipedWait.Minutes(), completed)
					a.logf("%s", msg)
					notifyMXU(ctx, msg)
				}
			}
			if wipedAutoRetreat && time.Since(wipedSince) >= wipedWait {
				msg := fmt.Sprintf("⚠️ 「全员重伤」已等待 %.0f 分钟，自动选择「全员撤退」并继续看护（本次已完成 %d 场）", time.Since(wipedSince).Minutes(), completed)
				a.logf("%s", msg)
				notifyMXU(ctx, msg)
				_, _ = ctx.RunTask("Battle_Coop_Auto_AllWiped")
				wipedSince = time.Time{}
				lastProgressTime = time.Now()
				if !sleepInterruptible(tasker, 3*time.Second) {
					return true
				}
				continue
			}
		} else if !wipedSince.IsZero() {
			a.logf("[Scan] 「全员重伤」弹窗已消失（队友已击杀或已结算），取消撤退计时")
			wipedSince = time.Time{}
		}

		// 4. Rising-edge detection on "战斗结算" screen ([40, 5, 200, 55]).
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_Settlement", img); err == nil && rec != nil && rec.Hit {
			if !inSettlement {
				inSettlement = true
				completed++
				lastProgressTime = time.Now()
				stage := ""
				if stageRec, sErr := ctx.RunRecognition("Battle_Coop_Auto_StageName", img); sErr == nil && stageRec != nil {
					stage = joinOCRText(stageRec.DetailJson, 0.5)
				}
				var msg string
				if stage != "" {
					msg = fmt.Sprintf("⚔️ 联机连战：已完成第 %d 场（%s）", completed, stage)
				} else {
					msg = fmt.Sprintf("⚔️ 联机连战：已完成第 %d 场", completed)
				}
				a.logf("%s", msg)
				notifyMXU(ctx, msg)
			}
		} else {
			inSettlement = false
		}

		// 5. Progress & soft-lock detection:
		// Check top-left battle HUD ("回合 X 时间 MM:SS").
		// Healthy battles keep advancing the timer; UI-disappeared soft-locks (boss animating with no HUD),
		// frozen-turn soft-locks, or infinite LOADING screens fail to advance HUD text.
		hudText := ""
		rawHudText := ""
		if rec, err := ctx.RunRecognition("Battle_Coop_Auto_BattleHud", img); err == nil && rec != nil {
			rawHudText = joinOCRText(rec.DetailJson, 0.4)
			if isBattleHudText(rawHudText) {
				hudText = rawHudText
			}
		}
		nowHudVisible := hudText != ""
		if nowHudVisible != hudVisible {
			hudVisible = nowHudVisible
			if hudVisible {
				a.logf("[Scan] 检测到战斗 HUD 出现: hud=%q completed=%d", hudText, completed)
			} else {
				a.logf("[Scan] 战斗 HUD 消失: raw=%q lastHud=%q completed=%d", rawHudText, lastHudText, completed)
			}
		}
		if hudText != "" && hudText != lastHudText {
			lastHudText = hudText
			lastProgressTime = time.Now()
		}

		stalled := time.Since(lastProgressTime)
		logInterval := 30 * time.Second
		if stalled >= 30*time.Second {
			logInterval = 15 * time.Second
		}
		if lastScanLogTime.IsZero() || time.Since(lastScanLogTime) >= logInterval {
			lastScanLogTime = time.Now()
			a.logf("[Scan] completed=%d inSettlement=%v hud=%q raw=%q stalled=%.0fs/%.0fs",
				completed, inSettlement, hudText, rawHudText, stalled.Seconds(), freezeTimeout.Seconds())
		}

		if freezeTimeout > 0 && stalled >= freezeTimeout {
			shot := a.saveDebugScreenshot(img, "freeze")
			var msg string
			if hudText == "" {
				msg = fmt.Sprintf("⚠️ 连续 %.0f 分钟未检测到战斗 HUD / 进度推进（疑似无UI软卡死或加载卡住），触发自动强杀重启续刷（本次已完成 %d 场）", freezeTimeout.Minutes(), completed)
			} else {
				msg = fmt.Sprintf("⚠️ 战斗计时连续 %.0f 分钟无变化（%s），触发自动强杀重启续刷（本次已完成 %d 场）", freezeTimeout.Minutes(), hudText, completed)
			}
			a.logf("%s [hud=%q raw=%q shot=%s]", msg, hudText, rawHudText, shot)
			notifyMXU(ctx, msg)
			_ = ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{{Name: "Battle_Coop_Auto_FreezeRecover"}})
			return true
		}

		if !sleepInterruptible(tasker, pollInterval) {
			return true
		}
	}
}

func runAgentServer(identifier string) int {
	root := rootDir()
	libDir := filepath.Join(root, "maafw")
	if err := maa.Init(maa.WithLibDir(libDir)); err != nil {
		fmt.Fprintf(os.Stderr, "[MET] maa.Init(%s) failed: %v\n", libDir, err)
		return 1
	}
	defer func() { _ = maa.Release() }()

	if err := maa.AgentServerRegisterCustomAction("CoopAutoMonitor", &CoopAutoMonitorAction{root: root}); err != nil {
		fmt.Fprintf(os.Stderr, "[MET] register CoopAutoMonitor failed: %v\n", err)
		return 1
	}

	if err := maa.AgentServerStartUp(identifier); err != nil {
		fmt.Fprintf(os.Stderr, "[MET] AgentServerStartUp(%s) failed: %v\n", identifier, err)
		return 1
	}
	maa.AgentServerJoin()
	maa.AgentServerShutDown()
	return 0
}
