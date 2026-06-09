TestXzRoundTrip() {
    printf 'roundtrip-xz\n' | xz | xzcat
}

TestLzmaRoundTrip() {
    printf 'roundtrip-lzma\n' | lzma | unlzma
}

TestZcatGzip() {
    d=$(mktemp -d)
    printf 'gz-payload\n' | gzip > "$d/f.gz"
    zcat "$d/f.gz"
    rm -rf "$d"
}

TestPipeProgress() {
    printf 'pass-through\n' | pipe_progress 2>/dev/null
}
