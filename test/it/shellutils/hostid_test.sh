TestHostidFormat() {
    hostid
}

TestHostidStable() {
    a=$(hostid)
    b=$(hostid)
    test "$a" = "$b" && echo "stable"
}
