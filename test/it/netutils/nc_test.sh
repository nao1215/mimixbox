Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestNcLoopback() {
    # Listener writes whatever the client sends into a file.
    (nc -l -p 18642 > /tmp/mimixbox/it/nc_recv.txt) &
    sleep 0.3
    echo "from-client" | nc 127.0.0.1 18642 >/dev/null 2>&1
    sleep 0.2
    cat /tmp/mimixbox/it/nc_recv.txt
}
