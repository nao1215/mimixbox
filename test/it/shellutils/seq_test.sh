TestSeqLast() {
    seq 3
}

TestSeqFirstLast() {
    seq 2 5
}

TestSeqFirstIncrementLast() {
    seq 1 2 9
}

TestSeqSeparator() {
    seq -s , 1 3
}

TestSeqEqualWidth() {
    seq -w 8 10
}

TestSeqInvalid() {
    seq abc
}
