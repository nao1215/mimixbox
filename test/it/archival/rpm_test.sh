# Fixture: test/it/testdata/sample.rpm (a minimal RPM with a gzip payload),
# package "hello-2.10-1.fc40.x86_64" containing /usr/bin/hello and
# /etc/hello.conf, payload "RPM-PAYLOAD".
fixture() {
    echo "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/testdata/sample.rpm"
}

TestRpmQuery() {
    rpm -qp "$(fixture)"
}
TestRpmList() {
    rpm -qpl "$(fixture)"
}
TestRpm2cpio() {
    rpm2cpio "$(fixture)"
}
