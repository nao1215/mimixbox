# shellcheck shell=sh
# Integration helper for `ipcalc`. See test/it/README.md.
#
# ipcalc is pure CIDR arithmetic with no network access, so no on-disk fixture
# is needed; the helper centralises representative deterministic invocations.

Setup() { export LANG=C; }
CleanUp() { :; }

# Full report for a /24 network.
TestIpcalcFull() {
    ipcalc 192.168.10.7/24
}

# Network address only.
TestIpcalcNetwork() {
    ipcalc -n 192.168.10.7/24
}

# Broadcast address only.
TestIpcalcBroadcast() {
    ipcalc -b 10.0.0.5/16
}
