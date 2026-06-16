# Dedicated integration helper for the 'scriptreplay' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ScriptreplayHelp() {
    env -- 'scriptreplay' --help
}
