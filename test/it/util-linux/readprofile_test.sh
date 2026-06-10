# /proc/profile is absent unless the kernel booted with profiling, so the e2e
# exercises --help; the buffer summary is covered by Go unit tests.
TestReadprofileHelp() { readprofile --help; }
