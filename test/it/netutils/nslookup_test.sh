# Dedicated integration helper for the 'nslookup' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
NslookupHelp() {
    env -- 'nslookup' --help
}
