# Logging Hook Script for GitHub Copilot Events
# Captures all Copilot interactions including messages, responses, and session details

param(
    [Parameter(Mandatory=$false)]
    [string]$Event = "stop"
)

# Get the script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogsDir = Join-Path $ScriptDir "../logs"

# Create logs directory if it doesn't exist
if (-not (Test-Path -Path $LogsDir)) {
    New-Item -ItemType Directory -Path $LogsDir -Force | Out-Null
}

# Log filename with today's date (yyyy-mm-dd format)
$LogDate = Get-Date -Format "yyyy-MM-dd"
$LogFile = Join-Path $LogsDir "$LogDate.log"
$ConversationFile = Join-Path $LogsDir "$LogDate-conversations.jsonl"

# Read stdin for payload (Copilot sends JSON data)
$InputData = @()
while ($input.MoveNext()) {
    $InputData += $input.Current
}
$Payload = $InputData -join "`n"

# Session details
$CurrentTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$Timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss.fff"
$Hostname = $env:COMPUTERNAME
$UserName = $env:USERNAME
$ShellVersion = "PowerShell $($PSVersionTable.PSVersion.Major).$($PSVersionTable.PSVersion.Minor)"
$WorkingDir = Get-Location
$ProcessId = $PID
$OperatingSystem = [System.Environment]::OSVersion.VersionString

# Try to get Git information
$GitBranch = "N/A"
$GitCommit = "N/A"
$GitStatus = "N/A"
try {
    $GitBranch = git rev-parse --abbrev-ref HEAD 2>$null
    $GitCommit = git rev-parse --short HEAD 2>$null
    $GitStatus = git status --short 2>$null
} catch {
    # Git not available or not a git repository
}

# Get available memory info (only for Stop event to avoid slowdown)
if ($Event -eq "stop") {
    $MemInfo = Get-CimInstance Win32_OperatingSystem | Select-Object @{Name="TotalMemory"; Expression={[math]::Round($_.TotalVisibleMemorySize / 1MB, 2)}}, @{Name="FreeMemory"; Expression={[math]::Round($_.FreePhysicalMemory / 1MB, 2)}}
    $MemoryInfo = "Memory: Total: $($MemInfo.TotalMemory) MB, Free: $($MemInfo.FreeMemory) MB"
} else {
    $MemoryInfo = "Memory: (skipped for performance)"
}

# Agent/CI Detection
$CIPlatform = "None"
$CIBuildId = "N/A"
$CIJobId = "N/A"
$CIAgentName = "N/A"
$ContainerInfo = "N/A"

# GitHub Actions
if ($env:GITHUB_ACTIONS -eq "true") {
    $CIPlatform = "GitHub Actions"
    $CIBuildId = $env:GITHUB_RUN_ID ?? "N/A"
    $CIJobId = $env:GITHUB_JOB ?? "N/A"
    $CIAgentName = $env:RUNNER_NAME ?? "N/A"
}

# GitLab CI
if ($env:GITLAB_CI -eq "true") {
    $CIPlatform = "GitLab CI"
    $CIBuildId = $env:CI_BUILD_ID ?? "N/A"
    $CIJobId = $env:CI_JOB_ID ?? "N/A"
    $CIAgentName = $env:CI_RUNNER_ID ?? "N/A"
}

# Azure Pipelines
if ($env:SYSTEM_TEAMFOUNDATIONCOLLECTIONURI) {
    $CIPlatform = "Azure Pipelines"
    $CIBuildId = $env:BUILD_BUILDID ?? "N/A"
    $CIJobId = $env:SYSTEM_JOBID ?? "N/A"
    $CIAgentName = $env:AGENT_NAME ?? "N/A"
}

# Docker Detection
if (Test-Path "/.dockerenv") {
    $ContainerInfo = "Docker Container (ID: $(hostname))"
} elseif (Test-Path "/proc/1/cgroup" -ErrorAction SilentlyContinue) {
    $CGroupContent = Get-Content "/proc/1/cgroup" -ErrorAction SilentlyContinue
    if ($CGroupContent -match "kubepods") {
        $ContainerInfo = "Kubernetes Pod"
    } elseif ($CGroupContent -match "lxc") {
        $ContainerInfo = "LXC Container"
    }
}

# Build structured log entry for JSONL
$StructuredLog = @{
    timestamp = $Timestamp
    event = $Event.ToUpper()
    hostname = $Hostname
    user = $UserName
    pid = $ProcessId
    shell = $ShellVersion
    workingDir = $WorkingDir.Path
    git = @{
        branch = $GitBranch
        commit = $GitCommit
        status = $GitStatus
    }
    ci = @{
        platform = $CIPlatform
        buildId = $CIBuildId
        jobId = $CIJobId
        agentName = $CIAgentName
        container = $ContainerInfo
    }
    payload = $Payload
}

# Save structured log to JSONL
$StructuredLog | ConvertTo-Json -Compress -Depth 10 | Add-Content -Path $ConversationFile -Encoding UTF8

# Separator for readability
$Separator = "================================================================================"

# Build human-readable log entry
$LogEntry = @"

$Separator
COPILOT EVENT: $($Event.ToUpper())
$Separator
Timestamp: $Timestamp
Date: $CurrentTime
Hostname: $Hostname
User: $UserName
PID: $ProcessId
Shell: $ShellVersion
Working Directory: $WorkingDir
OS Info: $OperatingSystem
$MemoryInfo

AGENT INFORMATION:
  CI Platform: $CIPlatform
  Build ID: $CIBuildId
  Job ID: $CIJobId
  Agent Name: $CIAgentName
  Container: $ContainerInfo

GIT INFORMATION:
  Branch: $GitBranch
  Commit: $GitCommit
  Status: $GitStatus

COPILOT PAYLOAD:
$Payload

$Separator

"@

# Append to log file
Add-Content -Path $LogFile -Value $LogEntry -Encoding UTF8

# Output for visibility
Write-Host "Copilot $Event event logged to: $LogFile"
Write-Host "Structured conversation data: $ConversationFile"

exit 0
