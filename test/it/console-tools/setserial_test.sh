# setserial parses serial parameters; applying them needs a real serial device.
# -g with explicit parameters echoes the parsed request (a deterministic, non
# no-op success path) without touching hardware.
TestSetserialGetEcho() { setserial -g /dev/ttyS0 baud_base 115200 | grep -c 'baud_base 115200'; }
# A bad parameter is rejected deterministically.
TestSetserialBadParam() { setserial /dev/ttyS0 bogus 1 2>/dev/null; echo "rc=$?"; }
# A missing device is rejected.
TestSetserialNoDevice() { setserial 2>/dev/null; echo "rc=$?"; }
