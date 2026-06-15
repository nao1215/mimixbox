# shellcheck shell=sh
# Integration helper for `http-status-code`. See test/it/README.md.
#
# Pure offline lookup table; no network. The helper centralises the search/list
# subcommand invocations.

Setup() { export LANG=C; }
CleanUp() { :; }

# Look up a single status code.
TestHttpStatusSearch() {
    http-status-code search 200
}

# Look up multiple status codes.
TestHttpStatusSearchMultiple() {
    http-status-code search 200 404 500
}

# List subcommand emits the full table.
TestHttpStatusList() {
    http-status-code list | head -n 1
}
