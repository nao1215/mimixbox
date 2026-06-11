TestRunPartsOrder() {
    d="$TEST_DIR/parts"; mkdir -p "$d"
    printf '#!/bin/sh\necho B\n' > "$d/20-b"; chmod +x "$d/20-b"
    printf '#!/bin/sh\necho A\n' > "$d/10-a"; chmod +x "$d/10-a"
    run-parts "$d"
}
TestRunPartsNoDir() { run-parts 2>/dev/null; echo "rc=$?"; }
