TestSum() { printf 'hello\n' | sum; }
TestCrc32() { printf 'hello\n' | crc32; }
TestSha384() { printf 'hello\n' | sha384sum; }

TestUuRoundTrip() {
    d=$(mktemp -d); printf 'round trip data\n' > "$d/in"
    uuencode "$d/in" out | uudecode -o -
    rm -rf "$d"
}

TestUuBase64() {
    d=$(mktemp -d); printf 'base64 round trip\n' > "$d/in"
    uuencode -m "$d/in" out | uudecode -o -
    rm -rf "$d"
}

TestUsleep() {
    usleep 1000 && echo slept
}

TestSha3Default() { printf 'hello\n' | sha3sum | cut -d' ' -f1; }
TestSha3_512() { printf 'hello\n' | sha3sum -a 512 | cut -c1-16; }
