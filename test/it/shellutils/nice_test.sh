# shellcheck shell=sh
# Integration helper for `nice`. See test/it/README.md.

Setup() { export LANG=C; }
CleanUp() { :; }

# With no command, nice prints the current niceness (an integer).
TestNiceReport() { nice; }

# nice runs the given command with an adjusted niceness; the command's own
# output passes through unchanged.
TestNiceRunCommand() { nice -n 5 echo nice-ok; }
