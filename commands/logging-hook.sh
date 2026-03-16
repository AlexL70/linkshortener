#!/bin/bash

# Logging Hook Script for GitHub Copilot Events
# Captures all Copilot interactions involving tool use (agent mode).
# Sessions with no tool use (ask/question mode) are silently discarded:
#   - SESSIONSTART and USERPROMPTSUBMITTED are buffered per session in /tmp.
#   - The buffer is flushed to the final log only when PRETOOLUSE fires.
#   - If SESSIONEND fires without any PRETOOLUSE, the buffer is discarded.

EVENT="${1:-stop}"
EVENT_UPPER=$(echo "$EVENT" | tr '[:lower:]' '[:upper:]')

# Get the logs directory relative to this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOGS_DIR="${SCRIPT_DIR}/../logs"

# Create logs directory if it doesn't exist
mkdir -p "$LOGS_DIR"

# Log filename with today's date (yyyy-mm-dd format)
LOG_DATE=$(date +%Y-%m-%d)
LOG_FILE="${LOGS_DIR}/${LOG_DATE}.log"
CONVERSATION_FILE="${LOGS_DIR}/${LOG_DATE}-conversations.jsonl"

# Read stdin for payload (Copilot sends JSON data)
PAYLOAD=$(cat)

# Session details
CURRENT_TIME=$(date '+%Y-%m-%d %H:%M:%S')
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S.%N')
HOSTNAME=$(hostname)
USER_NAME=$(whoami)
SHELL_TYPE="$SHELL"
PWD_DIR=$(pwd)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "N/A")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "N/A")
GIT_STATUS=$(git status --short 2>/dev/null | tr '\n' ' ' || echo "N/A")
UNAME=$(uname -a)
PROCESS_ID=$$

# Agent/CI Detection
CI_PLATFORM="None"
CI_BUILD_ID="N/A"
CI_JOB_ID="N/A"
CI_AGENT_NAME="N/A"
CONTAINER_INFO="N/A"

# GitHub Actions
if [ -n "$GITHUB_ACTIONS" ]; then
  CI_PLATFORM="GitHub Actions"
  CI_BUILD_ID="${GITHUB_RUN_ID:-N/A}"
  CI_JOB_ID="${GITHUB_JOB:-N/A}"
  CI_AGENT_NAME="${RUNNER_NAME:-N/A}"
fi

# GitLab CI
if [ -n "$GITLAB_CI" ]; then
  CI_PLATFORM="GitLab CI"
  CI_BUILD_ID="${CI_BUILD_ID:-N/A}"
  CI_JOB_ID="${CI_JOB_ID:-N/A}"
  CI_AGENT_NAME="${CI_RUNNER_ID:-N/A}"
fi

# Jenkins
if [ -n "$JENKINS_HOME" ]; then
  CI_PLATFORM="Jenkins"
  CI_BUILD_ID="${BUILD_ID:-N/A}"
  CI_JOB_ID="${BUILD_NUMBER:-N/A}"
  CI_AGENT_NAME="${NODE_NAME:-N/A}"
fi

# Docker Detection
if [ -f /.dockerenv ]; then
  CONTAINER_INFO="Docker Container (ID: $(hostname))"
elif grep -q "lxc" /proc/1/cgroup 2>/dev/null; then
  CONTAINER_INFO="LXC Container"
elif grep -q "kubepods" /proc/1/cgroup 2>/dev/null; then
  CONTAINER_INFO="Kubernetes Pod"
fi

# ---------------------------------------------------------------------------
# Per-session buffering setup
# ---------------------------------------------------------------------------
# Extract session_id from the payload (field name differs between event types).
SESSION_ID=$(echo "$PAYLOAD" | grep -oP '"session_id"\s*:\s*"\K[^"]+' | head -1)
if [ -z "$SESSION_ID" ]; then
  SESSION_ID=$(echo "$PAYLOAD" | grep -oP '"sessionId"\s*:\s*"\K[^"]+' | head -1)
fi

# Sanitise the ID so it is safe to use in a filename, then derive temp-file paths.
if [ -n "$SESSION_ID" ]; then
  SAFE_ID=$(echo "$SESSION_ID" | tr -cd 'a-zA-Z0-9_-' | head -c 64)
  JSONL_BUFFER="/tmp/copilot-hook-${SAFE_ID}.jsonl"
  LOG_BUFFER="/tmp/copilot-hook-${SAFE_ID}.log"
  # Presence of this file means the session has had at least one tool-use event.
  ACTIVE_MARKER="/tmp/copilot-hook-${SAFE_ID}.active"
else
  # Unknown session ID: fall back to direct logging so nothing is lost.
  JSONL_BUFFER=""
  LOG_BUFFER=""
  ACTIVE_MARKER=""
fi

# Separator for readability in the human-readable log
SEPARATOR="==============================================================================="

# ---------------------------------------------------------------------------
# Helper: append the current event to the given JSONL file and log file.
# Usage: append_event <dest_jsonl> <dest_log>
# ---------------------------------------------------------------------------
append_event() {
  local dest_jsonl="$1"
  local dest_log="$2"

  cat << EOJSON >> "$dest_jsonl"
{"timestamp":"${TIMESTAMP}","event":"${EVENT_UPPER}","hostname":"${HOSTNAME}","user":"${USER_NAME}","pid":${PROCESS_ID},"shell":"${SHELL_TYPE}","workingDir":"${PWD_DIR}","git":{"branch":"${GIT_BRANCH}","commit":"${GIT_COMMIT}","status":"${GIT_STATUS}"},"ci":{"platform":"${CI_PLATFORM}","buildId":"${CI_BUILD_ID}","jobId":"${CI_JOB_ID}","agentName":"${CI_AGENT_NAME}","container":"${CONTAINER_INFO}"},"payload":${PAYLOAD:-null}}
EOJSON

  local log_entry="
${SEPARATOR}
COPILOT EVENT: ${EVENT_UPPER}
${SEPARATOR}
Timestamp: ${TIMESTAMP}
Date: ${CURRENT_TIME}
Hostname: ${HOSTNAME}
User: ${USER_NAME}
PID: ${PROCESS_ID}
Shell: ${SHELL_TYPE}
Working Directory: ${PWD_DIR}
System Info: ${UNAME}

AGENT INFORMATION:
  CI Platform: ${CI_PLATFORM}
  Build ID: ${CI_BUILD_ID}
  Job ID: ${CI_JOB_ID}
  Agent Name: ${CI_AGENT_NAME}
  Container: ${CONTAINER_INFO}

GIT INFORMATION:
  Branch: ${GIT_BRANCH}
  Commit: ${GIT_COMMIT}
  Status: ${GIT_STATUS}

COPILOT PAYLOAD:
${PAYLOAD}

${SEPARATOR}
"
  echo "$log_entry" >> "$dest_log"
}

# ---------------------------------------------------------------------------
# Helper: flush session buffers to the final log files and mark session active.
# ---------------------------------------------------------------------------
flush_session_buffers() {
  [ -f "$JSONL_BUFFER" ] && cat "$JSONL_BUFFER" >> "$CONVERSATION_FILE" && rm -f "$JSONL_BUFFER"
  [ -f "$LOG_BUFFER"   ] && cat "$LOG_BUFFER"   >> "$LOG_FILE"          && rm -f "$LOG_BUFFER"
  touch "$ACTIVE_MARKER"
}

# ---------------------------------------------------------------------------
# Helper: remove all temp files for this session (called on SESSIONEND).
# ---------------------------------------------------------------------------
cleanup_session() {
  rm -f "$JSONL_BUFFER" "$LOG_BUFFER" "$ACTIVE_MARKER"
}

# ---------------------------------------------------------------------------
# Event routing
# ---------------------------------------------------------------------------
case "$EVENT_UPPER" in

  SESSIONSTART|USERPROMPTSUBMITTED)
    if [ -z "$ACTIVE_MARKER" ]; then
      # Unknown session ID — log directly so nothing is lost.
      append_event "$CONVERSATION_FILE" "$LOG_FILE"
    elif [ -f "$ACTIVE_MARKER" ]; then
      # Session already has tool use — this is an agent-mode session; log directly.
      append_event "$CONVERSATION_FILE" "$LOG_FILE"
    else
      # Session has not yet used any tools — buffer and wait.
      append_event "$JSONL_BUFFER" "$LOG_BUFFER"
    fi
    ;;

  PRETOOLUSE)
    # First tool use in this session: flush buffered events before logging this one.
    if [ -n "$ACTIVE_MARKER" ] && [ ! -f "$ACTIVE_MARKER" ]; then
      flush_session_buffers
    fi
    append_event "$CONVERSATION_FILE" "$LOG_FILE"
    ;;

  POSTTOOLUSE|AGENTSTOP)
    # Only log if the session has had tool use (PRETOOLUSE already fired).
    if [ -z "$ACTIVE_MARKER" ] || [ -f "$ACTIVE_MARKER" ]; then
      append_event "$CONVERSATION_FILE" "$LOG_FILE"
    fi
    ;;

  SESSIONEND)
    # Only log if the session had tool use; then clean up all temp files.
    if [ -z "$ACTIVE_MARKER" ] || [ -f "$ACTIVE_MARKER" ]; then
      append_event "$CONVERSATION_FILE" "$LOG_FILE"
    fi
    [ -n "$ACTIVE_MARKER" ] && cleanup_session
    ;;

  *)
    # Unknown / future events: log directly so nothing is silently lost.
    append_event "$CONVERSATION_FILE" "$LOG_FILE"
    ;;

esac

# Output for visibility
echo "Copilot ${EVENT} event processed (session: ${SESSION_ID:-unknown})."
echo "Log: ${LOG_FILE} | JSONL: ${CONVERSATION_FILE}"

exit 0
