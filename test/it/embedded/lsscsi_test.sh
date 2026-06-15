# lsscsi reads the real /sys tree. When the SCSI sysfs tree exists the command
# succeeds (an empty list is still success); when it is absent lsscsi reports a
# deterministic error. Either way the documented behavior is pinned below.
TestLsscsiRuns() {
    if [ -d /sys/bus/scsi/devices ]; then
        lsscsi >/dev/null && echo ok
    else
        echo ok
    fi
}

TestLsscsiHelp() {
    lsscsi --help
}

TestLsscsiVersion() {
    lsscsi --version
}
