TestDfHeader() {
    df . | head -n 1
}

TestDfStatus() {
    df . > /dev/null
    echo "rc=$?"
}
