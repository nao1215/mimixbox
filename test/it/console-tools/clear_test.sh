ClearHelp() { clear --help; }
# Discard the screen-clearing escape sequence; this contract test only asserts
# that clear exits successfully.
ClearRun() { clear >/dev/null; }
