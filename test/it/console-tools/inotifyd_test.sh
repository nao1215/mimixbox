TestInotifydWatchesCreate() {
    d="$TEST_DIR/w"; mkdir -p "$d"
    printf '#!/bin/sh\necho "$1 $3" >> %s/events\n' "$TEST_DIR" > "$TEST_DIR/h"
    chmod +x "$TEST_DIR/h"
    : > "$TEST_DIR/events"
    inotifyd "$TEST_DIR/h" "$d:n" &
    pid=$!
    # Give the watch time to be established, then create a file and poll for the
    # handler's output (more robust than fixed sleeps under CI load).
    sleep 0.3
    touch "$d/created.txt"
    for _ in $(seq 1 50); do
        if grep -q 'n created.txt' "$TEST_DIR/events" 2>/dev/null; then
            break
        fi
        sleep 0.1
    done
    kill "$pid" 2>/dev/null
    grep -c 'n created.txt' "$TEST_DIR/events"
}
TestInotifydNoArgs() { inotifyd ./h 2>/dev/null; echo "rc=$?"; }
