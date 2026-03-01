#!/bin/bash

# Logging Hook Script for GitHub Copilot Events
# Captures all Copilot interactions including messages, responses, and session details

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

# Build structured log entry for JSONL
cat << EOF >> "$CONVERSATION_FILE"
{"timestamp":"${TIMESTAMP}","event":"${EVENT_UPPER}","hostname":"${HOSTNAME}","user":"${USER_NAME}","pid":${PROCESS_ID},"shell":"${SHELL_TYPE}","workingDir":"${PWD_DIR}","git":{"branch":"${GIT_BRANCH}","commit":"${GIT_COMMIT}","status":"${GIT_STATUS}"},"ci":{"platform":"${CI_PLATFORM}","buildId":"${CI_BUILD_ID}","jobId":"${CI_JOB_ID}","agentName":"${CI_AGENT_NAME}","container":"${CONTAINER_INFO}"},"payload":${PAYLOAD:-null}}
EOF

# Separator for readability
SEPARATOR="==============================================================================="

# Build human-readable log entry
LOG_ENTRY="
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

# Append to log file
echo "$LOG_ENTRY" >> "$LOG_FILE"

# Also log to stdout for visibility
echo "Copilot ${EVENT} event logged to: ${LOG_FILE}"
echo "Structured conversation data: ${CONVERSATION_FILE}"

exit 0
