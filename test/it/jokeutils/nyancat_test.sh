TestNyancatNoTTY() {
    # No terminal under shellspec: exits gracefully (0) with no output.
    nyancat
    echo "rc:$?"
}
