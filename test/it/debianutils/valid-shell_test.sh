ValidShellHelp() { valid-shell --help; }
ValidShellValidFile() {
    f=$(mktemp)
    printf '/bin/sh\n/bin/bash\n' > "${f}"
    valid-shell "${f}"
    rc=$?
    rm -f "${f}"
    return ${rc}
}
