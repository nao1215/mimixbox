# syslogd runs until interrupted, so the e2e exercises --help; the socket-to-log
# behavior is covered by a Go integration-style unit test.
TestSyslogdHelp() { syslogd --help; }
