Setup() {
    export TEST_DIR=/tmp/mimixbox/it
    mkdir -p ${TEST_DIR}
    printf 'alpha\nhunter2\nbeta\n' > ${TEST_DIR}/words.txt
    printf 'UEsDBAoACQAAAPeJyFzf7VBXIAAAABQAAAAIABwAZGF0YS50eHRVVAkAAzF6JmoxeiZqdXgLAAEE6AMAAAToAwAAEbWyZmDC+DuePSekzjycZ7l/UpFToIZ+b2pmJIB3fahQSwcI3+1QVyAAAAAUAAAAUEsBAh4DCgAJAAAA94nIXN/tUFcgAAAAFAAAAAgAGAAAAAAAAQAAAKSBAAAAAGRhdGEudHh0VVQFAAMxeiZqdXgLAAEE6AMAAAToAwAAUEsFBgAAAAABAAEATgAAAHIAAAAAAA==' | base64 -d > ${TEST_DIR}/enc.zip
}
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestZipPwcrack() {
    zip-pwcrack /tmp/mimixbox/it/enc.zip -w /tmp/mimixbox/it/words.txt
}
