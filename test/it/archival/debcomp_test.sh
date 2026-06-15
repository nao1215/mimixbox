TestBzip2RoundTrip() {
    printf 'roundtrip-bzip2\n' | bzip2 | bzip2 -dc
}

TestLzopRoundTrip() {
    printf 'roundtrip-lzop\n' | lzop | lzopcat
}

TestUnlzopRoundTrip() {
    printf 'roundtrip-unlzop\n' | lzop | unlzop -c
}

# Fixture: test/it/testdata/hello.deb (a minimal Debian package, format 2.0),
# package "hello" version 1.0 containing /usr/bin/hello. It is a committed
# binary fixture rather than built at runtime, because under the E2E PATH
# `tar`/`ar` resolve to the MimixBox applets and cannot assemble a .deb with the
# dashless flag bundles GNU tooling accepts.
deb_fixture() {
    echo "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/testdata/hello.deb"
}

TestDpkgDebContents() {
    dpkg-deb -c "$(deb_fixture)" | grep -q 'usr/bin/hello' && printf 'has-hello\n'
}

TestDpkgDebField() {
    dpkg-deb -f "$(deb_fixture)" Package
}

TestDpkgExtract() {
    d=$(mktemp -d)
    out="$d/out"
    dpkg -x "$(deb_fixture)" "$out"
    test -f "$out/usr/bin/hello" && printf 'extracted\n'
    rm -rf "$d"
}

TestDpkgUnsupported() {
    if dpkg -i "$(deb_fixture)" 2>/dev/null; then
        printf 'unexpected-success\n'
    else
        printf 'rejected\n'
    fi
}
