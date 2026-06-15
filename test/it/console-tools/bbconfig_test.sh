# bbconfig prints the build configuration of the running MimixBox binary.
TestBbconfigHasVersionLine() { bbconfig | grep -c 'CONFIG_MIMIXBOX_VERSION='; }
# Every applet line is CONFIG_<NAME>=y; bbconfig must list itself.
TestBbconfigListsItself() { bbconfig --names | grep -c '^bbconfig$'; }
# An unexpected operand is a deterministic error.
TestBbconfigRejectsArg() { bbconfig extra 2>/dev/null; echo "rc=$?"; }
