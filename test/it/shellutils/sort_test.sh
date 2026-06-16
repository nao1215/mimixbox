TestSortLexical() {
    printf 'banana\napple\ncherry\n' | sort
}

TestSortNumeric() {
    printf '10\n2\n1\n' | sort -n
}

TestSortReverse() {
    printf 'a\nb\nc\n' | sort -r
}

TestSortUnique() {
    printf 'a\na\nb\n' | sort -u
}

# TestSortVersion checks -V orders embedded version numbers by value so that
# 1.2 precedes 1.10 (unlike a plain lexical sort).
TestSortVersion() {
    printf '1.10\n1.2\n1.1\n' | sort -V
}

# TestSortGeneralNumeric checks -g compares floating-point values, including
# scientific notation that plain -n would not fully parse.
TestSortGeneralNumeric() {
    printf '1e3\n2.5\n100\n0.5\n' | sort -g
}

# TestSortHumanNumeric checks -h orders human-readable sizes by magnitude so
# that 2K < 1M < 1G.
TestSortHumanNumeric() {
    printf '1G\n2K\n1M\n' | sort -h
}

# TestSortStable checks -s keeps the input order of lines whose numeric key is
# equal rather than breaking the tie with a full-line comparison.
TestSortStable() {
    printf '5 zebra\n5 apple\n5 mango\n' | sort -s -n
}

# TestSortZeroTerminated checks -z reads and writes NUL-delimited records; the
# boundaries are rendered as '|' so the result is comparable as plain text.
TestSortZeroTerminated() {
    printf 'banana\000apple\000cherry\000' | sort -z | tr '\000' '|'
}

# TestSortMerge checks -m produces a sorted result from already-sorted input.
TestSortMerge() {
    printf 'apple\nbanana\ncherry\n' | sort -m
}

# TestSortParallel checks --parallel=N is accepted and does not error.
TestSortParallel() {
    printf 'b\na\n' | sort --parallel=4
}

# TestSortTemporaryDirectory checks --temporary-directory=DIR is accepted and
# does not error.
TestSortTemporaryDirectory() {
    printf 'b\na\n' | sort --temporary-directory=/tmp
}
