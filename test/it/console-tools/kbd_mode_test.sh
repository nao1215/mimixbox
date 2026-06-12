# Reading/setting the keyboard mode needs a console; the e2e exercises the
# deterministic conflicting-options path and --help. The ioctls are unit-tested.
TestKbdModeConflict() { kbd_mode -a -u 2>/dev/null; echo "rc=$?"; }
TestKbdModeHelp() { kbd_mode --help; }
