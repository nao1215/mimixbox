Setup() {
    export TEST_DIR=/tmp/mimixbox/it/blkid
    mkdir -p ${TEST_DIR}
    # Craft an image with the ext superblock magic (0xEF53 at offset 1080).
    dd if=/dev/zero of=${TEST_DIR}/ext.img bs=1024 count=2 2>/dev/null
    printf '\123\357' | dd of=${TEST_DIR}/ext.img bs=1 seek=1080 conv=notrunc 2>/dev/null
    printf 'XFSB' > ${TEST_DIR}/xfs.img
    printf 'nothing here' > ${TEST_DIR}/blank.img
}

CleanUp() { rm -rf /tmp/mimixbox/it/blkid; }

TestBlkidExt() { blkid ${TEST_DIR}/ext.img; }
TestBlkidXfs() { blkid ${TEST_DIR}/xfs.img; }
TestBlkidNone() { blkid ${TEST_DIR}/blank.img 2>/dev/null; echo "rc=$?"; }
