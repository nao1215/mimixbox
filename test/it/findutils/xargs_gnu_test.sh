# -L groups input by line: -L 1 runs the command once per input line.
TestXargsMaxLinesOne() {
    printf 'a b\nc d\ne f\n' | xargs -L 1 echo | wc -l | tr -d ' '
}

# -L 2 groups two input lines per invocation: 3 lines -> 2 invocations.
TestXargsMaxLinesTwo() {
    printf 'a\nb\nc\n' | xargs -L 2 echo | wc -l | tr -d ' '
}

# -s limits the command line length, splitting one long input into several
# invocations. With an 8-char budget the eight single-char items cannot all fit
# on one line, so more than one invocation is produced.
TestXargsMaxCharsSplits() {
    printf '1 2 3 4 5 6 7 8\n' | xargs -s 8 echo | wc -l | tr -d ' '
}

# -s must still emit every item exactly once across the split invocations.
TestXargsMaxCharsKeepsAllItems() {
    printf '1 2 3 4 5 6 7 8\n' | xargs -s 8 echo | tr ' ' '\n' | grep -c .
}

# -P runs invocations concurrently; all batches still run, so all four outputs
# appear regardless of order. Sort to make the unordered result deterministic.
TestXargsMaxProcsAllRun() {
    printf 'a\nb\nc\nd\n' | xargs -P 4 -n 1 echo | sort | tr '\n' ' ' | sed 's/ $//'
}

# -P 0 means "run as many as possible"; every batch still completes.
TestXargsMaxProcsZero() {
    printf 'a\nb\nc\nd\n' | xargs -P 0 -n 1 echo | sort | tr '\n' ' ' | sed 's/ $//'
}
