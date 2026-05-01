# wtf — Windows installer (PowerShell).
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/kitay-sudo/wtf/main/install.ps1 | iex
#
# Закрепить версию: $env:WTF_VERSION = "v0.1.0"  перед запуском.

$ErrorActionPreference = "Stop"

$Repo = "kitay-sudo/wtf"
$BinName = "wtf.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\wtf"

function Step($msg) { Write-Host "  → $msg" -ForegroundColor Cyan }
function OK($msg)   { Write-Host "  ✓ $msg" -ForegroundColor Yellow }
function Info($msg) { Write-Host "  ⓘ $msg" -ForegroundColor DarkGray }
function Err($msg)  { Write-Host "  ✗ $msg" -ForegroundColor Red; exit 1 }

# ---- detect arch ----
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Err "32-битная Windows не поддерживается"
}

# ---- latest version ----
$version = if ($env:WTF_VERSION) { $env:WTF_VERSION } else {
    try {
        (Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest").tag_name
    } catch {
        # 404 = у репо нет релизов.
        if ($_.Exception.Response.StatusCode.value__ -eq 404) {
            Write-Host ""
            Write-Host "  ✗ У репозитория $Repo ещё нет ни одного релиза." -ForegroundColor Red
            Write-Host ""
            Write-Host "  Маинтейнеру: выпусти первый релиз —"
            Write-Host "    scripts\release.bat v0.1.0"
            Write-Host ""
            Write-Host "  После того как workflow release.yml завершится, эта команда заработает."
            Write-Host ""
            exit 1
        }
        Err "не удалось получить latest version с GitHub: $_"
    }
}

Write-Host ""
Write-Host "  [!?] wtf installer" -NoNewline -ForegroundColor White
Write-Host " · " -NoNewline
Write-Host $version -NoNewline -ForegroundColor Yellow
Write-Host " · windows/$arch"
Write-Host ""

# ---- download ----
$url = "https://github.com/$Repo/releases/download/$version/wtf_windows_${arch}.zip"
$tmp = Join-Path $env:TEMP "wtf-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmp | Out-Null

Step "скачиваю $url"
try {
    Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmp "wtf.zip") -UseBasicParsing
} catch {
    Err "не удалось скачать: $_"
}
OK "скачано"

Step "распаковываю"
Expand-Archive -Path (Join-Path $tmp "wtf.zip") -DestinationPath $tmp -Force
$srcBin = Join-Path $tmp $BinName
if (-not (Test-Path $srcBin)) { Err "в архиве нет $BinName" }
OK "распаковано"

# ---- install ----
Step "устанавливаю в $InstallDir"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}
Copy-Item -Path $srcBin -Destination (Join-Path $InstallDir $BinName) -Force
OK "установлено"

# ---- PATH ----
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    Step "добавляю $InstallDir в PATH (User)"
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    OK "PATH обновлён — перезапусти терминал чтобы изменения применились"
} else {
    Info "PATH уже содержит $InstallDir"
}

# ---- cleanup ----
Remove-Item -Recurse -Force $tmp

Write-Host ""
Write-Host "  [!?] готово!" -ForegroundColor White
Write-Host ""
Write-Host "  Дальше:"
Write-Host "    PS> wtf config              " -NoNewline; Write-Host "# настроить провайдера и ключ" -ForegroundColor DarkGray
Write-Host "    PS> wtf nginx не стартует   " -NoNewline; Write-Host "# запустить диагностику" -ForegroundColor DarkGray
Write-Host "    PS> wtf memory show         " -NoNewline; Write-Host "# что агент о тебе помнит" -ForegroundColor DarkGray
Write-Host ""
