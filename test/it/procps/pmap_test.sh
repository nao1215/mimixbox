TestPmapTotal() { pmap $$ | tail -n 1 | grep -c 'total'; }
TestPmapInvalid() { pmap notapid 2>/dev/null; echo "rc=$?"; }
