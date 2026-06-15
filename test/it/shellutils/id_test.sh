# shellcheck shell=sh
# Integration helper for `id`. See test/it/README.md.
#
# id reads the current process identity; output is environment-dependent, so
# Test functions assert on stable structure (numeric ids, name lookups) rather
# than absolute values.

Setup() { export LANG=C; }
CleanUp() { :; }

# Default form includes uid=, gid= and groups= fields.
TestIdDefault() { id; }

# Effective user id (numeric).
TestIdUserNumeric() { id -u; }

# Effective user name.
TestIdUserName() { id -un; }

# Effective group id (numeric).
TestIdGroupNumeric() { id -g; }
