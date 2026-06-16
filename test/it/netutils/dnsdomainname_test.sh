# Dedicated integration helper for the 'dnsdomainname' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
DnsdomainnameHelp() {
    env -- 'dnsdomainname' --help
}
