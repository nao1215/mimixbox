# The tree connectors are multibyte; grep for an ASCII node label to confirm the
# tree was built. PID 1 (init/systemd) is always present.
TestPstreeHasInit() {
    pstree | grep -c '(1)'
}
