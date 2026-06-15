# shellcheck shell=sh
# Integration helper for `getopt`. See test/it/README.md.
#
# getopt is pure argument parsing (no fixtures on disk), but a dedicated helper
# centralises the common normalised invocation so specs do not re-derive it.

Setup() { export LANG=C; }
CleanUp() { :; }

# Enhanced mode: short options only, normalised output.
TestGetoptShort() {
    getopt -o ab: -- -a -b x file
}

# Enhanced mode with long options.
TestGetoptLong() {
    getopt -o a --long alpha,beta: -- -a --beta=1 operand
}
