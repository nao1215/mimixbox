# Dedicated integration helper for the 'ifenslave' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IfenslaveHelp() {
    env -- 'ifenslave' --help
}
