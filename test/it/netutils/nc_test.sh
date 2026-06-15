Setup() { export TEST_DIR=${MIMIXBOX_IT_ROOT}; mkdir -p ${TEST_DIR}; }
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}; }

# NcSocketsForbidden is a Skip-predicate: it succeeds (status 0) when the host
# refuses to open the loopback listen socket the test needs. It briefly starts a
# listener and inspects its stderr for a capability error (e.g. "operation not
# permitted" / "permission denied"). It deliberately returns failure (sockets
# are usable) on any ambiguous result so a real nc regression is never masked.
NcSocketsForbidden() {
    probe_dir=${MIMIXBOX_IT_ROOT:-${TMPDIR:-/tmp}}
    mkdir -p "$probe_dir" 2>/dev/null
    probe_err="$probe_dir/nc_probe.err"
    : > "$probe_err"

    (nc -l -p 18639 >/dev/null 2>"$probe_err") &
    lpid=$!
    sleep 0.3
    kill "$lpid" 2>/dev/null
    wait "$lpid" 2>/dev/null

    grep -qiE 'operation not permitted|permission denied|not permitted|socket:' "$probe_err"
}
TestNcLoopback() {
    recv=${MIMIXBOX_IT_ROOT}/nc_recv.txt
    : > "$recv"

    # Each outer attempt starts a fresh listener on a fresh port and then retries
    # the client send. Restarting the listener (rather than only retrying the
    # client) is what makes this robust on loaded CI: the observed flake was an
    # empty result because the one-shot listener occasionally failed to bind in
    # time, and a fresh port also dodges TIME_WAIT from a previous attempt. The
    # loop keys off the received file's contents, so it does not depend on the
    # nc client's or listener's exit status.
    for attempt in $(seq 1 8); do
        port=$((18640 + attempt))
        (nc -l -p "$port" > "$recv") &
        lpid=$!

        for _ in $(seq 1 20); do
            echo "from-client" | nc 127.0.0.1 "$port" >/dev/null 2>&1
            if [ -s "$recv" ]; then
                break
            fi
            sleep 0.1
        done

        kill "$lpid" 2>/dev/null
        wait "$lpid" 2>/dev/null
        if [ -s "$recv" ]; then
            break
        fi
    done

    # Print only the first received line: if a retry races and two client sends
    # both land before the listener closes, the file would otherwise contain the
    # message twice and break the exact-match assertion.
    head -n 1 "$recv"
}
