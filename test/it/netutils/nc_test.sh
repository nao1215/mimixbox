Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestNcLoopback() {
    # Listener writes whatever the client sends into a file.
    (nc -l -p 18642 > /tmp/mimixbox/it/nc_recv.txt) &

    # Retry the client send until the listener has bound and recorded the
    # message, instead of relying on fixed sleeps. Fixed sleeps race against the
    # listener's bind/flush under CI load and made this test flaky. The loop
    # keys off the received file's contents, so it is robust regardless of the
    # nc client's exit status.
    for _ in $(seq 1 50); do
        echo "from-client" | nc 127.0.0.1 18642 >/dev/null 2>&1
        if [ -s /tmp/mimixbox/it/nc_recv.txt ]; then
            break
        fi
        sleep 0.1
    done

    cat /tmp/mimixbox/it/nc_recv.txt
}
