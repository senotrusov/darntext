#!/usr/bin/env bash

# Copyright 2025-2026 Stanislav Senotrusov
#
# This work is dual-licensed under the Apache License, Version 2.0
# and the MIT License. Refer to the LICENSE file in the top-level directory
# for the full license terms.
#
# SPDX-License-Identifier: Apache-2.0 OR MIT

# dir: ~/example

# Deduplicate gettext translation messages.
#
# Uses Expo's msguniq Mix task to rewrite each domain file in place.
#
deduplicate_locale_po_files() {
  local po_file

  # Match all .po files within any locale's LC_MESSAGES directory
  for po_file in priv/gettext/*/LC_MESSAGES/*.po; do
    mix expo.msguniq "$po_file" --output-file "$po_file" &
  done
  wait
}

# Files that should be added to the generated project context.
context_cmd() {
  files=(
    assets/css/*.css
    assets/js/*.js
    assets/tsconfig.json
    config/**/*.exs
    lib/**/*.ex*
    lib/**/*.heex
    priv/gettext/**/*.po

    .tool-versions
    ./*.exs
    README.md

    # test/**/*.ex*
  )

  add_files "${files[@]}"

  db_schema

  migration_datetime

  agents_md

  add_prompt general-guidelines
  # add_prompt testing
  add_prompt translations
}

apply_cmd() {
  # Apply incoming changes, appending new entries to PO files.
  term-paste | darntext-apply -append-po

  # Remove duplicate gettext entries introduced during the apply step.
  deduplicate_locale_po_files

  # Format code and merge extracted gettext messages after applying updates.
  tidy_cmd
}

tidy_cmd() {
  # Format Elixir source and other files handled by mix format.
  mix format

  # Extract gettext strings and merge them into existing PO/POT files.
  mix gettext.extract --merge
}

default_cmd() {
  echo apply
  echo context
  echo re
  echo task
  echo tidy
}

invoke_handler "$@"
