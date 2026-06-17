#!/usr/bin/env bash

# Copyright 2025-2026 Stanislav Senotrusov
#
# This work is dual-licensed under the Apache License, Version 2.0
# and the MIT License. Refer to the LICENSE file in the top-level directory
# for the full license terms.
#
# SPDX-License-Identifier: Apache-2.0 OR MIT

# Exit on errors and treat unset variables as errors
set -eu

# Set an ERR trap to report errors and enable propagation of ERR traps through
# functions, subshells, and command substitutions.
trap 'echo "Error: The command \"$BASH_COMMAND\" at ${FUNCNAME:-main} \
($BASH_SOURCE:$LINENO) failed with exit status $?" >&2' ERR; set -E

# Make pipelines fail if any command in the pipeline fails
set -o pipefail

# Enable recursive globbing (**) and ignore non-matching patterns
shopt -s globstar nullglob

# ----

STATE_DIR="${XDG_STATE_HOME:-"${HOME}/.local/state"}/darntext/$(systemd-escape -p "$PWD")"

invoke_handler() {
  # Match pattern: alphanumeric and hyphens only (no underscores)
  if [[ "${1:-}" =~ ^[a-zA-Z0-9-]+$ ]] && declare -F "$1_cmd" >/dev/null 2>&1; then
    "$1_cmd" "${@:2}"
  elif declare -F "default_cmd" >/dev/null 2>&1; then
    default_cmd "$@"
  else
    echo "Error: command not provided and default command is not defined." >&2
    return 1
  fi
}

make_state() {
  local timestamp

  mkdir -p "$STATE_DIR"
  find "$STATE_DIR" -type f -mtime +30 -delete &

  timestamp="$(date "+%Y-%m-%dT%H:%M:%S")"
  mktemp "${STATE_DIR}/${timestamp}-XXX${1:+-$1}"
}

edit() {
  if [ -n "${EDITOR:-}" ] && command -v "$EDITOR" >/dev/null 2>&1; then
    "$EDITOR" "$1"
    return
  fi

  local editor
  for editor in editor micro nano vim vi; do
    if command -v "$editor" >/dev/null 2>&1; then
      "$editor" "$1"
      return
    fi
  done

  printf 'Error: no editor found\n' >&2
  return 1
}

add_files() {
  darntext-add-files "$@"
}

db_schema() {
  darntext-db-schema "$@"
}

add_prompt() {
  cat "${DARNTEXT_CONFIG_DIR}/prompts/$1.md"
}

agents_md() {
  # Include AGENTS.md raw if it exists in the current directory
  if [ -f "AGENTS.md" ]; then
    printf '\n'
    cat "AGENTS.md"
  fi
}

migration_datetime() {
  echo "* If you need to create a database migration, use this $(date -u +"%Y%m%d%H%M%S") datetime"
  echo "  as part of the filename."
}

run_task() {
  local task="$1"
  shift

  edit "$task"

  wait # need to wait to not interfere with osc-52 output

  {
    context_cmd "$@"
    if [ -s "$task" ]; then
      printf '\n# Objective\n\n'
      cat "$task"
    fi
  } | term-copy
  
  printf '\nTask copied to a clipboard!\n' >&2
}

task_cmd() {
  local task
  task="$(make_state task)"

  run_task "$task" "$@"
}

re_cmd() {
  local files
  local selected
  local task
  local fzf_status

  if ! command -v fzf >/dev/null 2>&1; then
    echo "Error: fzf is not installed or not in PATH" >&2
    return 1
  fi
  
  # Collect previous tasks ending in -task or -tasks.
  # Shell globbing sorts them alphabetically, placing the most recent timestamp at the end.
  files=("${STATE_DIR}"/*-task)

  if [ ${#files[@]} -eq 0 ]; then
    echo "Error: No past tasks found in ${STATE_DIR}" >&2
    return 1
  fi

  # Use '/' as a delimiter and select the last field (-1) to display only the filename.
  # This keeps the interface clean while preserving the full path for the selection output and preview.
  fzf_status=0
  selected="$(printf '%s\n' "${files[@]}" | tac | fzf --delimiter '/' --with-nth -1 --preview 'cat {}' --preview-window 'right,50%')" || fzf_status=$?

  if [ -z "$selected" ] || [ "$fzf_status" = 130 ]; then
    return 0
  fi

  if [ "$fzf_status" != 0 ]; then
    return 1
  fi

  task="$(make_state task)"
  cp "$selected" "$task"

  run_task "$task" "$@"
}
