# Dedicated integration helper for the 'route' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
RouteHelp() {
    env -- 'route' --help
}
