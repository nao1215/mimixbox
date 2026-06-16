# Dedicated integration helper for the 'resume' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ResumeHelp() {
    env -- 'resume' --help
}
