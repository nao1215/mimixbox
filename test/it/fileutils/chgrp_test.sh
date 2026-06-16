# Dedicated integration helper for the 'chgrp' command (GitHub issue #489).
# Exercises the structured --help contract through the installed applet.
ChgrpHelp() {
    env -- 'chgrp' --help
}
