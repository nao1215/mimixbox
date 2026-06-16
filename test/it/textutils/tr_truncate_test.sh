# --truncate-set1 cuts SET1 (abc) to the length of SET2 (xy), so a->x, b->y and
# c is left unchanged, yielding "xyc".
TestTrTruncateSet1() { printf 'abc\n' | tr --truncate-set1 abc xy; }

# The -t short form behaves identically.
TestTrTruncateSet1Short() { printf 'abc\n' | tr -t abc xy; }
