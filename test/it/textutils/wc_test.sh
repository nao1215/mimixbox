

Setup() {
    # local test_dir=/tmp/mimixbox/it
    # local test_file=/tmp/mimixbox/it/test.txt

    mkdir -p /tmp/mimixbox/it
    echo "NieR Replicant ver.1.22474487139..." > /tmp/mimixbox/it/test.txt
    echo "NieR:Automata" >> /tmp/mimixbox/it/test.txt
    echo "The Legend of Zelda: Majora's Mask" >> /tmp/mimixbox/it/test.txt
    echo "KICHIKUOU RANCE" >> /tmp/mimixbox/it/test.txt
    echo "DARK SOULS" >> /tmp/mimixbox/it/test.txt
    echo "SHADOW HEARTS" >> /tmp/mimixbox/it/test.txt
}

CleanUp() {
    #local test_file=/tmp/mimixbox/it/test.txt
    rm /tmp/mimixbox/it/test.txt
}

TestWcWithNoOption() {
    #local test_file=/tmp/mimixbox/it/test.txt
    mimixbox wc /tmp/mimixbox/it/test.txt
}