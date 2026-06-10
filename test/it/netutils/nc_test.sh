Setup() { export TEST_DIR=/tmp/mimixbox/it; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestNcLoopback() {
    # Listener writes whatever the client sends into a file.
    (nc -l -p 18642 > /tmp/mimixbox/it/nc_recv.txt) &

    # Retry the client send until the listener has bound and recorded the
    # message, instead of relying on fixed sleeps. Fixed sleeps race against the
    # listener's bind/flush under CI load and made this test flaky. The loop
    # keys off the received file's contents, so it is robust regardless of the
    # nc client's exit status. The window is generous (10s) because the failure
    # mode under heavy CI load is the listener not having bound in time.
    for _ in $(seq 1 100); do
        echo "from-client" | nc 127.0.0.1 18642 >/dev/null 2>&1
        if [ -s /tmp/mimixbox/it/nc_recv.txt ]; then
            break
        fi
        sleep 0.1
    done

    # Print only the first received line: if a retry races and two client sends
    # both land before the listener closes, the file would otherwise contain the
    # message twice and break the exact-match assertion.
    head -n 1 /tmp/mimixbox/it/nc_recv.txt
}
