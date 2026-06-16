# Dedicated integration helper for the 'ipneigh' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
IpneighHelp() {
    env -- 'ipneigh' --help
}
