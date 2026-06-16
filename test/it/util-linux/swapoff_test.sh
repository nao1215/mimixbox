# Dedicated integration helper for the 'swapoff' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SwapoffHelp() {
    env -- 'swapoff' --help
}
