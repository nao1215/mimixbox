# killall5 signals every process, so the e2e only exercises the non-destructive
# --help path; the signal dispatch is covered by Go unit tests.
TestKillall5Help() { killall5 --help; }
