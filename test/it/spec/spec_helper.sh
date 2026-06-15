# shellcheck shell=sh

# --- Issue #478: per-run integration-test temp root -------------------------
# Historically the suite hard-coded /tmp/mimixbox/it as a shared, global temp
# root. That broke `make it` whenever /tmp/mimixbox already existed as a file
# or symlink, and made parallel/repeated local runs share mutable state.
#
# Instead the suite now uses a single per-run root exported as MIMIXBOX_IT_ROOT.
# When invoked through `make it`/`make test-e2e`, the Makefile allocates it via
# `mktemp -d` and removes it after the run. When shellspec is run directly
# (e.g. `cd test/it && shellspec`), allocate a fallback here so the suite stays
# runnable standalone. Exporting it makes the value reach every example subshell
# and every Included <category>/<name>_test.sh helper.
if [ -z "${MIMIXBOX_IT_ROOT:-}" ]; then
  MIMIXBOX_IT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/mimixbox.XXXXXX")
fi
export MIMIXBOX_IT_ROOT

# it_root prints the per-run temp root. It is a convenience alias for use inside
# spec assertions via command substitution, e.g.
#   The output should equal "$(it_root)/find/a.txt"
# Equivalent to referencing "${MIMIXBOX_IT_ROOT}" directly.
it_root() { printf '%s' "${MIMIXBOX_IT_ROOT}"; }

# Defining variables and functions here will affect all specfiles.
# Change shell options inside a function may cause different behavior,
# so it is better to set them here.
# set -eu

# This callback function will be invoked only once before loading specfiles.
spec_helper_precheck() {
  # Available functions: info, warn, error, abort, setenv, unsetenv
  # Available variables: VERSION, SHELL_TYPE, SHELL_VERSION
  : minimum_version "0.28.1"
}

# This callback function will be invoked after a specfile has been loaded.
spec_helper_loaded() {
  :
}

# This callback function will be invoked after core modules has been loaded.
spec_helper_configure() {
  # Available functions: import, before_each, after_each, before_all, after_all
  : import 'support/custom_matcher'
}
