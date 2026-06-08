Setup() { export D=/tmp/mimixbox/it; mkdir -p $D; printf 'hello gzip roundtrip\n' > $D/g.txt; }
CleanUp() { rm -rf /tmp/mimixbox/it; }
TestGzipRoundtrip() {
    gzip /tmp/mimixbox/it/g.txt
    gunzip /tmp/mimixbox/it/g.txt.gz
    cat /tmp/mimixbox/it/g.txt
}
