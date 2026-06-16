# Dedicated integration helper for the 'crc32' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
Crc32Help() {
    env -- 'crc32' --help
}
