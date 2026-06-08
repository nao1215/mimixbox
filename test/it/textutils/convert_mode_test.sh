Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'a\r\nb\r\n' > ${TEST_DIR}/d2u.txt
    chmod 600 ${TEST_DIR}/d2u.txt
    printf 'a\nb\n' > ${TEST_DIR}/u2d.txt
    chmod 600 ${TEST_DIR}/u2d.txt
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestDos2unixKeepsMode() {
    dos2unix ${TEST_DIR}/d2u.txt >/dev/null 2>&1
    stat -c '%a' ${TEST_DIR}/d2u.txt
}
TestUnix2dosKeepsMode() {
    unix2dos ${TEST_DIR}/u2d.txt >/dev/null 2>&1
    stat -c '%a' ${TEST_DIR}/u2d.txt
}
