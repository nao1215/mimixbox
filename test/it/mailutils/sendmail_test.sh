# Dedicated integration helper for the 'sendmail' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
SendmailHelp() {
    env -- 'sendmail' --help
}
