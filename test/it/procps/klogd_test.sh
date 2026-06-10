# Forwarding to syslog needs a running syslogd and kernel-buffer access, so the
# e2e exercises --help; the parsing and forwarding are covered by unit tests.
TestKlogdHelp() { klogd --help; }
