TestMkfsReiserRefuses() { mkfs.reiser /tmp/x.img 2>/dev/null; echo "rc=$?"; }
TestMkfsReiserMessage() { mkfs.reiser /tmp/x.img 2>&1 | grep -c 'deprecated'; }
