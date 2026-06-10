TestLsblkHeader() { lsblk | sed -n '1p'; }
TestLsblkRuns() { lsblk >/dev/null 2>&1; echo $?; }
