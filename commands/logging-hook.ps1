# Logging Hook Script for GitHub Copilot Events
# Buffers SESSIONSTART and USERPROMPTSUBMITTED events per session.
# Events are only written to the final log when tool use is detected
# (PRETOOLUSE), ensuring pure ask/question-mode sessions produce no log output.

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

# ---------------------------------------------------------------------------
# Per-session buffering setup
# ---------------------------------------------------------------------------
# Extract session_id from the payload (field name differs between event types).
$SessionId = $null
if ($Payload) {
    if ($Payload -match '"session_id"\s*:\s*"([^"]+)"') {
        $SessionId = $Matches[1]
    } elseif ($Payload -match '"sessionId"\s*:\s*"([^"]+)"') {
        $SessionId = $Matches[1]
    }
}

# Sanitise the ID so it is safe to use in a filename, then derive temp-file paths.
if ($SessionId) {
    $SafeId = ($SessionId -replace '[^a-zA-Z0-9_-]', '') 
    if ($SafeId.Length -gt 64) { $SafeId = $SafeId.Substring(0, 64) }
    $JsonlBuffer  = Join-Path $env:TEMP "copilot-hook-$SafeId.jsonl"
    $LogBuffer    = Join-Path $env:TEMP "copilot-hook-$SafeId.log"
    # Presence of this file means the session has had at least one tool-use event.
    $ActiveMarker = Join-Path $env:TEMP "copilot-hook-$SafeId.active"
} else {
    # Unknown session ID: fall back to direct logging so nothing is lost.
    $JsonlBuffer  = $null
    $LogBuffer    = $null
    $ActiveMarker = $null
}

# Separator for readability
$Separator = "================================================================================"

# ---------------------------------------------------------------------------
# Helper: build the structured JSONL object for this event.
# ---------------------------------------------------------------------------
function Build-StructuredLog {
    return @{
        timestamp = $Timestamp
        event     = $Event.ToUpper()
        hostname  = $Hostname
        user      = $UserName
        pid       = $ProcessId
        shell     = $ShellVersion
        workingDir = $WorkingDir.Path
        git = @{
            branch = $GitBranch
            commit = $GitCommit
            status = $GitStatus
        }
        ci = @{
            platform  = $CIPlatform
            buildId   = $CIBuildId
            jobId     = $CIJobId
            agentName = $CIAgentName
            container = $ContainerInfo
        }
        payload = $Payload
    }
}

# ---------------------------------------------------------------------------
# Helper: build the human-readable log entry for this event.
# ---------------------------------------------------------------------------
function Build-LogEntry {
    return @"

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
}

# ---------------------------------------------------------------------------
# Helper: append the current event to the supplied destination files.
# ---------------------------------------------------------------------------
function Append-Event {
    param([string]$DestJsonl, [string]$DestLog)
    (Build-StructuredLog) | ConvertTo-Json -Compress -Depth 10 | Add-Content -Path $DestJsonl -Encoding UTF8
    Add-Content -Path $DestLog -Value (Build-LogEntry) -Encoding UTF8
}

# ---------------------------------------------------------------------------
# Helper: flush session buffers to the final log files and mark session active.
# ---------------------------------------------------------------------------
function Flush-SessionBuffers {
    if (Test-Path $JsonlBuffer) {
        Get-Content $JsonlBuffer | Add-Content -Path $ConversationFile -Encoding UTF8
        Remove-Item $JsonlBuffer -Force
    }
    if (Test-Path $LogBuffer) {
        Get-Content $LogBuffer | Add-Content -Path $LogFile -Encoding UTF8
        Remove-Item $LogBuffer -Force
    }
    New-Item -ItemType File -Path $ActiveMarker -Force | Out-Null
}

# ---------------------------------------------------------------------------
# Helper: remove all temp files for this session (called on SESSIONEND).
# ---------------------------------------------------------------------------
function Cleanup-Session {
    foreach ($f in @($JsonlBuffer, $LogBuffer, $ActiveMarker)) {
        if ($f -and (Test-Path $f)) { Remove-Item $f -Force }
    }
}

# ---------------------------------------------------------------------------
# Event routing
# ---------------------------------------------------------------------------
$EventUpper = $Event.ToUpper()

switch ($EventUpper) {

    { $_ -in @("SESSIONSTART", "USERPROMPTSUBMITTED") } {
        if (-not $ActiveMarker) {
            # Unknown session ID — log directly so nothing is lost.
            Append-Event $ConversationFile $LogFile
        } elseif (Test-Path $ActiveMarker) {
            # Session already has tool use — log directly.
            Append-Event $ConversationFile $LogFile
        } else {
            # Session has not yet used any tools — buffer and wait.
            Append-Event $JsonlBuffer $LogBuffer
        }
    }

    "PRETOOLUSE" {
        # First tool use in this session: flush buffered events before logging this one.
        if ($ActiveMarker -and -not (Test-Path $ActiveMarker)) {
            Flush-SessionBuffers
        }
        Append-Event $ConversationFile $LogFile
    }

    { $_ -in @("POSTTOOLUSE", "AGENTSTOP") } {
        # Only log if the session has had tool use (PRETOOLUSE already fired).
        if (-not $ActiveMarker -or (Test-Path $ActiveMarker)) {
            Append-Event $ConversationFile $LogFile
        }
    }

    "SESSIONEND" {
        # Only log if the session had tool use; then clean up all temp files.
        if (-not $ActiveMarker -or (Test-Path $ActiveMarker)) {
            Append-Event $ConversationFile $LogFile
        }
        if ($ActiveMarker) { Cleanup-Session }
    }

    default {
        # Unknown / future events: log directly so nothing is silently lost.
        Append-Event $ConversationFile $LogFile
    }
}

# Output for visibility
Write-Host "Copilot $Event event processed (session: $($SessionId ?? 'unknown'))."
Write-Host "Log: $LogFile | JSONL: $ConversationFile"

exit 0
