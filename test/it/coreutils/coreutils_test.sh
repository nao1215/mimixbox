TestFactor() { factor 360; }
TestTsort() { printf 'a b\nb c\n' | tsort; }
TestEgrep() { printf 'foo\nbar\nbaz\n' | egrep 'ba(r|z)'; }
TestFgrep() { printf 'a.b\naxb\n' | fgrep 'a.b'; }
