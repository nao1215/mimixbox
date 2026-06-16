# Dedicated integration helper for the 'ssl_client' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SslClientHelp() {
    env -- 'ssl_client' --help
}
