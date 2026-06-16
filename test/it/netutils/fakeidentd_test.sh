# Dedicated integration helper for the 'fakeidentd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
FakeidentdHelp() {
    env -- 'fakeidentd' --help
}
