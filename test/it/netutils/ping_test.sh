TestPingUsage() {
    # No HOST: usage error (does not need raw-socket privileges).
    ping 2>&1
    echo "rc:$?"
}
