param(
    [Parameter(Position = 0)]
    [string]$OptionsJson = ""
)

$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

function Get-GameProcess {
    return Get-Process -Name "eternaltree" -ErrorAction SilentlyContinue | Select-Object -First 1
}

function Find-GameExeAuto {
    $globs = @(
        "C:\TapTap\PC Games\176733\*\eternaltree.exe",
        "D:\TapTap\PC Games\176733\*\eternaltree.exe",
        "E:\TapTap\PC Games\176733\*\eternaltree.exe",
        "F:\TapTap\PC Games\176733\*\eternaltree.exe"
    )
    foreach ($pattern in $globs) {
        $match = Get-Item -Path $pattern -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($null -ne $match) {
            return $match.FullName
        }
    }

    $regRoots = @(
        "HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*",
        "HKLM:\SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*",
        "HKCU:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*"
    )
    foreach ($root in $regRoots) {
        $items = Get-ItemProperty -Path $root -ErrorAction SilentlyContinue | Where-Object {
            $_.DisplayName -match "(悠久之树|eternaltree|Eternal\s*Tree)"
        }
        foreach ($item in $items) {
            if (-not [string]::IsNullOrWhiteSpace($item.InstallLocation)) {
                $candidate = Join-Path $item.InstallLocation "eternaltree.exe"
                if (Test-Path -LiteralPath $candidate) {
                    return $candidate
                }
            }
        }
    }
    return $null
}

# 1. Process already running => immediately succeed (zero C# compile overhead).
$existing = Get-GameProcess
if ($null -ne $existing) {
    Write-Host "[MET] 检测到 eternaltree.exe 已在运行 (PID=$($existing.Id))，跳过启动。"
    exit 0
}

# 2. Parse configured exe path from MXU pretask option JSON (if provided).
$pathMode = "Auto"
$exePath = ""
if (-not [string]::IsNullOrWhiteSpace($OptionsJson)) {
    try {
        $parsed = $OptionsJson | ConvertFrom-Json
        if ($null -ne $parsed.PCGamePathMode -and -not [string]::IsNullOrWhiteSpace($parsed.PCGamePathMode)) {
            $pathMode = [string]$parsed.PCGamePathMode
        }
        if ($null -ne $parsed.PCGameExePath -and -not [string]::IsNullOrWhiteSpace($parsed.PCGameExePath.GameExe)) {
            $exePath = $parsed.PCGameExePath.GameExe.Trim().Trim('"')
        }
    } catch {
        Write-Warning "[MET] 解析 pretask 参数失败，将尝试自动探测游戏路径: $($_.Exception.Message)"
    }
}

# 3. Auto-detect when mode is Auto (or manual path left blank).
if ($pathMode -eq "Auto" -or [string]::IsNullOrWhiteSpace($exePath)) {
    $detected = Find-GameExeAuto
    if (-not [string]::IsNullOrWhiteSpace($detected)) {
        $exePath = $detected
        Write-Host "[MET] 自动探测到游戏路径: $exePath"
    }
}

if ([string]::IsNullOrWhiteSpace($exePath) -or -not (Test-Path -LiteralPath $exePath)) {
    Write-Error "[MET] 未检测到运行中的 eternaltree.exe，且未找到有效的游戏可执行文件路径。请在「启动 eternaltree.exe」中配置路径，或先手动启动游戏。"
    exit 1
}

Write-Host "[MET] 正在启动游戏: $exePath"
$workDir = Split-Path -Parent $exePath
Start-Process -FilePath $exePath -WorkingDirectory $workDir | Out-Null

# 4. Wait up to 60s for the Unity main window to appear.
$deadline = (Get-Date).AddSeconds(60)
while ((Get-Date) -lt $deadline) {
    $proc = Get-GameProcess
    if ($null -ne $proc) {
        try { $proc.Refresh() } catch {}
        if ($proc.MainWindowHandle -ne [IntPtr]::Zero) {
            Start-Sleep -Seconds 2
            Write-Host "[MET] 游戏窗口已就绪 (PID=$($proc.Id))。"
            exit 0
        }
    }
    Start-Sleep -Milliseconds 500
}

Write-Error "[MET] 启动游戏超时（60秒内未检测到 eternaltree 主窗口）。"
exit 1
