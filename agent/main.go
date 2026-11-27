//go:build windows

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procFindWindowW              = user32.NewProc("FindWindowW")
	procIsIconic                 = user32.NewProc("IsIconic")
	procShowWindow               = user32.NewProc("ShowWindow")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procTerminateProcess         = kernel32.NewProc("TerminateProcess")
)

const (
	th32csSnapProcess = 0x00000002
	processTerminate  = 0x0001
	swRestore         = 9
)

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

type pretaskOptions struct {
	PCGamePathMode string `json:"PCGamePathMode"`
	PCGameExePath  struct {
		GameExe string `json:"GameExe"`
	} `json:"PCGameExePath"`
}

type mxuConfig struct {
	LastActiveInstanceID string `json:"lastActiveInstanceId"`
	Settings             struct {
		WebServerEnabled *bool `json:"webServerEnabled"`
		WebServerPort    int   `json:"webServerPort"`
	} `json:"settings"`
	Instances []struct {
		ID          string `json:"id"`
		SavedDevice struct {
			ConnectedProgramPath string `json:"connectedProgramPath"`
		} `json:"savedDevice"`
	} `json:"instances"`
}

func findProcessPIDs(targetName string) []uint32 {
	snap, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snap == 0 || snap == ^uintptr(0) {
		return nil
	}
	defer syscall.CloseHandle(syscall.Handle(snap))

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	r, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&entry)))
	if r == 0 {
		return nil
	}

	var pids []uint32
	for {
		name := syscall.UTF16ToString(entry.ExeFile[:])
		if strings.EqualFold(name, targetName) {
			pids = append(pids, entry.ProcessID)
		}
		r, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&entry)))
		if r == 0 {
			break
		}
	}
	return pids
}

func killProcess(pid uint32) error {
	h, _, err := procOpenProcess.Call(processTerminate, 0, uintptr(pid))
	if h == 0 {
		return err
	}
	defer syscall.CloseHandle(syscall.Handle(h))
	r, _, err := procTerminateProcess.Call(h, 1)
	if r == 0 {
		return err
	}
	return nil
}

func findGameWindow() uintptr {
	cls, _ := syscall.UTF16PtrFromString("UnityWndClass")
	for _, title := range []string{"eternaltree", "悠久之树"} {
		t, _ := syscall.UTF16PtrFromString(title)
		hwnd, _, _ := procFindWindowW.Call(uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(t)))
		if hwnd != 0 {
			return hwnd
		}
	}
	return 0
}

func restoreIfMinimized(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	iconic, _, _ := procIsIconic.Call(hwnd)
	if iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
		time.Sleep(300 * time.Millisecond)
	}
}

func rootDir() string {
	exe, err := os.Executable()
	if err != nil {
		cwd, _ := os.Getwd()
		return cwd
	}
	// agent/met-service.exe -> parent of agent/
	dir := filepath.Dir(exe)
	if strings.EqualFold(filepath.Base(dir), "agent") {
		return filepath.Dir(dir)
	}
	return dir
}

func loadMXUConfig(root string) *mxuConfig {
	data, err := os.ReadFile(filepath.Join(root, "config", "mxu-MET.json"))
	if err != nil {
		return nil
	}
	var cfg mxuConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	return &cfg
}

func findGameExeAuto(root string) string {
	// 1. Check savedDevice.connectedProgramPath in config/mxu-MET.json
	if cfg := loadMXUConfig(root); cfg != nil {
		for _, inst := range cfg.Instances {
			p := strings.TrimSpace(inst.SavedDevice.ConnectedProgramPath)
			if p != "" {
				if _, err := os.Stat(p); err == nil {
					return p
				}
			}
		}
	}
	// 2. Check common TapTap install locations
	for _, drive := range []string{"C", "D", "E", "F"} {
		pattern := fmt.Sprintf(`%s:\TapTap\PC Games\176733\*\eternaltree.exe`, drive)
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

func launchExe(exePath string) error {
	cmd := exec.Command(exePath)
	cmd.Dir = filepath.Dir(exePath)
	return cmd.Start()
}

func runPretaskLaunchPC(optJSON string) int {
	root := rootDir()

	// 1. If eternaltree.exe is already running, restore if minimized and exit 0 immediately.
	if pids := findProcessPIDs("eternaltree.exe"); len(pids) > 0 {
		hwnd := findGameWindow()
		restoreIfMinimized(hwnd)
		fmt.Printf("[MET] 检测到 eternaltree.exe 已在运行 (PID=%d)，跳过启动。\n", pids[0])
		return 0
	}

	// 2. Parse optional pretask JSON from MXU.
	pathMode := "Auto"
	exePath := ""
	if strings.TrimSpace(optJSON) != "" {
		var opts pretaskOptions
		if err := json.Unmarshal([]byte(optJSON), &opts); err == nil {
			if opts.PCGamePathMode != "" {
				pathMode = opts.PCGamePathMode
			}
			exePath = strings.Trim(strings.TrimSpace(opts.PCGameExePath.GameExe), `"`)
		}
	}

	// 3. Auto-detect when mode is Auto or path empty.
	if pathMode == "Auto" || exePath == "" {
		if detected := findGameExeAuto(root); detected != "" {
			exePath = detected
			fmt.Printf("[MET] 自动探测到游戏路径: %s\n", exePath)
		}
	}

	if exePath == "" {
		fmt.Fprintln(os.Stderr, "[MET] 未检测到运行中的 eternaltree.exe，请在「启动 eternaltree.exe」中配置路径或先手动启动游戏。")
		return 1
	}
	if _, err := os.Stat(exePath); err != nil {
		fmt.Fprintf(os.Stderr, "[MET] 游戏路径不存在: %s\n", exePath)
		return 1
	}

	fmt.Printf("[MET] 正在启动游戏: %s\n", exePath)
	if err := launchExe(exePath); err != nil {
		fmt.Fprintf(os.Stderr, "[MET] 启动失败: %v\n", err)
		return 1
	}

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if hwnd := findGameWindow(); hwnd != 0 {
			restoreIfMinimized(hwnd)
			time.Sleep(2 * time.Second)
			fmt.Printf("[MET] 游戏窗口已就绪 (hwnd=%d)。\n", hwnd)
			return 0
		}
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Fprintln(os.Stderr, "[MET] 启动游戏超时（60秒内未检测到 eternaltree 主窗口）。")
	return 1
}

func runRestartCoopPC() int {
	root := rootDir()
	logDir := filepath.Join(root, "debug")
	_ = os.MkdirAll(logDir, 0o755)
	logPath := filepath.Join(logDir, "met_coop_watchdog.log")

	logf := func(format string, args ...any) {
		msg := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			_, _ = f.WriteString(msg)
			_ = f.Close()
		}
	}

	logf("检测到联机自动战斗画面卡死，开始执行 PC 客户端强杀与重启续刷流程...")
	time.Sleep(2 * time.Second)

	cfg := loadMXUConfig(root)
	instanceID := ""
	port := 12701
	webEnabled := true
	if cfg != nil {
		instanceID = cfg.LastActiveInstanceID
		if cfg.Settings.WebServerPort > 0 {
			port = cfg.Settings.WebServerPort
		}
		if cfg.Settings.WebServerEnabled != nil {
			webEnabled = *cfg.Settings.WebServerEnabled
		}
		if instanceID == "" && len(cfg.Instances) > 0 {
			instanceID = cfg.Instances[0].ID
		}
	}

	for _, pid := range findProcessPIDs("eternaltree.exe") {
		logf("强杀卡死进程 eternaltree.exe (PID=%d)", pid)
		_ = killProcess(pid)
	}
	time.Sleep(2 * time.Second)

	exePath := findGameExeAuto(root)
	if exePath == "" {
		logf("错误: 未找到 eternaltree.exe 路径，无法自动重启。")
		return 1
	}

	logf("正在重新启动游戏: %s", exePath)
	if err := launchExe(exePath); err != nil {
		logf("错误: 启动游戏失败: %v", err)
		return 1
	}

	var newHwnd uint64
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if hwnd := findGameWindow(); hwnd != 0 {
			restoreIfMinimized(hwnd)
			newHwnd = uint64(hwnd)
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if newHwnd == 0 {
		logf("错误: 60秒内未检测到新启动的 eternaltree 窗口。")
		return 1
	}

	logf("检测到新游戏窗口 hwnd=%d，等待 3 秒初始化...", newHwnd)
	time.Sleep(3 * time.Second)

	if !webEnabled || instanceID == "" {
		logf("提示: MXU 本地 Web 服务未启用或未找到实例 ID，已完成客户端重启。")
		return 0
	}

	client := &http.Client{Timeout: 10 * time.Second}
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	connectPayload, _ := json.Marshal(map[string]any{
		"type":               "Win32",
		"handle":             newHwnd,
		"screencap_method":   2, // FramePool
		"mouse_method":       1, // Seize
		"keyboard_method":    1, // Seize
		"display_short_side": 720,
	})
	resp, err := client.Post(fmt.Sprintf("%s/api/maa/instances/%s/connect", baseURL, instanceID), "application/json", bytes.NewReader(connectPayload))
	if err != nil {
		logf("重连 MXU 失败: %v", err)
		return 1
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	logf("重连响应 (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	if resp.StatusCode >= 300 {
		return 1
	}

	time.Sleep(1 * time.Second)
	resPayload, _ := json.Marshal(map[string]any{
		"paths": []string{
			filepath.Join(root, "resource", "base"),
			filepath.Join(root, "resource", "win32"),
		},
	})
	if resResp, resErr := client.Post(fmt.Sprintf("%s/api/maa/instances/%s/resource/load", baseURL, instanceID), "application/json", bytes.NewReader(resPayload)); resErr == nil {
		resBody, _ := io.ReadAll(resResp.Body)
		_ = resResp.Body.Close()
		logf("重载资源响应 (%d): %s", resResp.StatusCode, strings.TrimSpace(string(resBody)))
	}
	time.Sleep(1 * time.Second)
	tasksPayload, _ := json.Marshal(map[string]any{
		"tasks": []map[string]string{
			{"entry": "Startup_Launch", "pipeline_override": "{}"},
			{"entry": "Battle_Coop_Auto_Start", "pipeline_override": "{}"},
		},
		"agent_configs": []map[string]any{
			{
				"child_exec": "agent/met-service.exe",
				"child_args": []string{"--agent"},
			},
		},
	})
	resp2, err := client.Post(fmt.Sprintf("%s/api/maa/instances/%s/tasks/start", baseURL, instanceID), "application/json", bytes.NewReader(tasksPayload))
	if err != nil {
		logf("派发续刷任务失败: %v", err)
		return 1
	}
	body2, _ := io.ReadAll(resp2.Body)
	_ = resp2.Body.Close()
	logf("续刷响应 (%d): %s", resp2.StatusCode, strings.TrimSpace(string(body2)))
	return 0
}

func main() {
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "--pretask" {
		name := args[1]
		optJSON := ""
		if len(args) >= 3 {
			optJSON = args[len(args)-1]
		}
		if name == "LaunchGamePC" {
			os.Exit(runPretaskLaunchPC(optJSON))
		}
	}
	if len(args) >= 1 && args[0] == "--restart-coop" {
		os.Exit(runRestartCoopPC())
	}
	if len(args) >= 2 && args[0] == "--agent" {
		os.Exit(runAgentServer(args[len(args)-1]))
	}
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		os.Exit(runAgentServer(args[0]))
	}
	fmt.Fprintln(os.Stderr, "Usage: met-service --pretask LaunchGamePC [optionsJSON] | --restart-coop | --agent <identifier>")
	os.Exit(2)
}
