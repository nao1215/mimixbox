# Renice this shell to its current niceness (a no-op change that needs no extra
# privilege), and confirm the GNU-style report line.
TestRenice() {
    renice -n 0 -p $$ | grep -c 'process ID'
}

TestReniceInvalid() {
    renice 5 notapid 2>/dev/null
    echo "rc=$?"
}
