TestPidofInit() {
    # Invoke MimixBox's own pidof explicitly so the test can never silently
    # validate the host system's pidof (GitHub issue #265).
    mimixbox pidof mimixbox >/dev/null 2>&1
    # pidof of a definitely-running process: the test shell's own 'sleep'
    sleep 5 &
    SLEEP_PID=$!
    RESULT=$(mimixbox pidof sleep)
    kill ${SLEEP_PID} 2>/dev/null
    echo "${RESULT}" | grep -q "${SLEEP_PID}" && echo found
}

# TestPidofIsMimixBox asserts that the bare 'pidof' on PATH resolves to a
# MimixBox-installed symlink, not the host binary. After 'mimixbox
# --full-install', the symlink must exist and point at the mimixbox binary.
TestPidofIsMimixBox() {
    pidof_path=$(command -v pidof) || return 1
    [ -L "${pidof_path}" ] || return 1
    target=$(readlink "${pidof_path}")
    case "${target}" in
        *mimixbox) echo linked ;;
        *) return 1 ;;
    esac
}
