# BangmodMonitor Agent — Windows installer
# Usage: irm https://your-domain.com/install.ps1 | iex
# Or:    .\install.ps1 -Token <TOKEN> [-Server URL] [-Region th] [-Interval 30]

param(
  [Parameter(Mandatory=$true)]
  [string]$Token,

  [string]$Server   = "https://api.bangmodmonitor.com",
  [string]$Region   = "default",
  [int]   $Interval = 30
)

$ErrorActionPreference = "Stop"
$ServiceName  = "BangmodAgent"
$InstallDir   = "$env:ProgramFiles\BangmodMonitor"
$BinaryPath   = "$InstallDir\bangmod-agent.exe"
$DownloadUrl  = "$Server/downloads/bangmod-agent-windows-amd64.exe"

Write-Host "==> Installing BangmodMonitor Agent" -ForegroundColor Cyan
Write-Host "    Server : $Server"
Write-Host "    Region : $Region"

# Create install directory
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null

# Download agent binary
Write-Host "==> Downloading agent..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $BinaryPath -UseBasicParsing

# Remove existing service if present
$existing = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existing) {
  Write-Host "==> Stopping existing service..."
  Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
  sc.exe delete $ServiceName | Out-Null
  Start-Sleep -Seconds 2
}

# Build environment variable string for the service
# Using a wrapper script to set env vars since sc.exe doesn't support them directly
$WrapperPath = "$InstallDir\start-agent.ps1"
@"
`$env:AGENT_TOKEN   = '$Token'
`$env:API_URL       = '$Server'
`$env:AGENT_REGION  = '$Region'
`$env:AGENT_INTERVAL = '$Interval'
& '$BinaryPath'
"@ | Set-Content -Path $WrapperPath -Encoding UTF8

# Create Windows Service via sc.exe
$BinPathCmd = "powershell.exe -NonInteractive -NoProfile -ExecutionPolicy Bypass -File `"$WrapperPath`""
sc.exe create $ServiceName binPath= $BinPathCmd start= auto | Out-Null
sc.exe description $ServiceName "BangmodMonitor Agent — collects system metrics" | Out-Null
sc.exe failure $ServiceName reset= 60 actions= restart/10000/restart/10000/restart/30000 | Out-Null

# Start the service
Start-Service -Name $ServiceName
$svc = Get-Service -Name $ServiceName
Write-Host ""
Write-Host "==> Service status: $($svc.Status)" -ForegroundColor Green
Write-Host "==> BangmodMonitor Agent installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "    Manage service:"
Write-Host "      Start   : Start-Service $ServiceName"
Write-Host "      Stop    : Stop-Service $ServiceName"
Write-Host "      Status  : Get-Service $ServiceName"
Write-Host "      Logs    : Get-EventLog -LogName Application -Source $ServiceName -Newest 20"
