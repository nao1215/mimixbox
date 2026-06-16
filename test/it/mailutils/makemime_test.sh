# Dedicated integration helper for the 'makemime' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
MakemimeHelp() {
    env -- 'makemime' --help
}
