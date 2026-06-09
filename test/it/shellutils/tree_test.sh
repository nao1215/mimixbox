# The tree body uses multibyte box-drawing characters; shellspec's output
# capture cannot carry those through the hermetic toolchain, and the exact
# formatting is already covered by the Go unit tests. Here we assert on the
# ASCII summary line, which confirms the traversal counted every entry.
TestTreeSummary() {
    d=$(mktemp -d)
    mkdir -p "$d/sub"
    touch "$d/sub/leaf.txt" "$d/root.txt"
    tree "$d" | grep directories
    rm -rf "$d"
}

TestTreeStatus() {
    d=$(mktemp -d)
    touch "$d/f.txt"
    tree "$d" > /dev/null
    status=$?
    rm -rf "$d"
    echo "$status"
}

TestNicePrints() {
    nice
}
