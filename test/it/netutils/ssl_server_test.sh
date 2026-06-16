# Dedicated integration helper for the 'ssl_server' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SslServerHelp() {
    env -- 'ssl_server' --help
}
