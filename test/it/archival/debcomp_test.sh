TestBzip2RoundTrip() {
    printf 'roundtrip-bzip2\n' | bzip2 | bzip2 -dc
}

TestLzopRoundTrip() {
    printf 'roundtrip-lzop\n' | lzop | lzopcat
}

TestUnlzopRoundTrip() {
    printf 'roundtrip-unlzop\n' | lzop | unlzop -c
}

# build_deb DIR creates a minimal hello.deb in DIR using mimixbox applets and
# prints its path.
build_deb() {
    d="$1"
    mkdir -p "$d/work/usr/bin"
    printf '#!/bin/sh\necho hi\n' > "$d/work/usr/bin/hello"
    mkdir -p "$d/ctrl"
    printf 'Package: hello\nVersion: 1.0\nArchitecture: all\nDescription: test\n' > "$d/ctrl/control"

    ( cd "$d/ctrl" && tar czf "$d/control.tar.gz" control )
    ( cd "$d/work" && tar czf "$d/data.tar.gz" usr )
    printf '2.0\n' > "$d/debian-binary"
    ( cd "$d" && ar rc hello.deb debian-binary control.tar.gz data.tar.gz )
    printf '%s/hello.deb\n' "$d"
}

TestDpkgDebContents() {
    d=$(mktemp -d)
    deb=$(build_deb "$d")
    dpkg-deb -c "$deb" | grep -q 'usr/bin/hello' && printf 'has-hello\n'
    rm -rf "$d"
}

TestDpkgDebField() {
    d=$(mktemp -d)
    deb=$(build_deb "$d")
    dpkg-deb -f "$deb" Package
    rm -rf "$d"
}

TestDpkgExtract() {
    d=$(mktemp -d)
    deb=$(build_deb "$d")
    out="$d/out"
    dpkg -x "$deb" "$out"
    test -f "$out/usr/bin/hello" && printf 'extracted\n'
    rm -rf "$d"
}

TestDpkgUnsupported() {
    d=$(mktemp -d)
    deb=$(build_deb "$d")
    if dpkg -i "$deb" 2>/dev/null; then
        printf 'unexpected-success\n'
    else
        printf 'rejected\n'
    fi
    rm -rf "$d"
}
