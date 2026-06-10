# Delivering to syslog needs a running syslogd, so the e2e exercises the
# deterministic priority-validation path; logging itself is covered by unit tests.
TestLoggerBadPriority() {
    logger -p nosuchfacility.info msg 2>/dev/null
    echo "rc=$?"
}
