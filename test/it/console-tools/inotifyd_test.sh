TestInotifydWatchesCreate() {
    d="$TEST_DIR/w"; mkdir -p "$d"
    printf '#!/bin/sh\necho "$1 $3" >> %s/events\n' "$TEST_DIR" > "$TEST_DIR/h"
    chmod +x "$TEST_DIR/h"
    : > "$TEST_DIR/events"
    inotifyd "$TEST_DIR/h" "$d:n" &
    pid=$!
    # Recreate the watched file each iteration until the handler logs the create
    # event, so a create lost before the watch is fully active is retried.
    found=missing
    for _ in $(seq 1 50); do
        rm -f "$d/created.txt"
        touch "$d/created.txt"
        if grep -q 'n created.txt' "$TEST_DIR/events" 2>/dev/null; then
            found=ok
            break
        fi
        sleep 0.1
    done
    kill "$pid" 2>/dev/null
    wait "$pid" 2>/dev/null
    echo "$found"
}
TestInotifydNoArgs() { inotifyd ./h 2>/dev/null; echo "rc=$?"; }
