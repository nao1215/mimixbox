# Dedicated integration helper for the 'httpd' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
HttpdHelp() {
    env -- 'httpd' --help
}
