# Dedicated integration helper for the 'dpkg-deb' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DpkgDebHelp() {
    env -- 'dpkg-deb' --help
}
