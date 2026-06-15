# shellcheck shell=sh
# Integration helper for `groups`. See test/it/README.md.

Setup() { export LANG=C; }
CleanUp() { :; }

# Current user's groups (space-separated names).
TestGroupsCurrent() { groups; }
