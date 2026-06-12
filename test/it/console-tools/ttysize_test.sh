# Run without a controlling tty, so ttysize reports the 80x24 default.
TestTtysize() { ttysize </dev/null; }
TestTtysizeWidth() { ttysize w </dev/null; }
