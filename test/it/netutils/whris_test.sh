TestWhrisUsage() {
    whris 2>&1
    echo "rc:$?"
}
