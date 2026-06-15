Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/makedevs
    mkdir -p ${TEST_DIR}
    cat > ${TEST_DIR}/table.txt <<'EOF'
# device table
/dev d 755 0 0 0 0 0 0 0
/etc/hostname f 644 0 0 0 0 0 0 0
EOF
}
CleanUp() { rm -rf ${MIMIXBOX_IT_ROOT}/makedevs; }

# Builds the tree and reports 1 when both the directory and file exist.
TestMakedevsTree() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/makedevs
    makedevs -d ${TEST_DIR}/table.txt ${TEST_DIR}/rootfs
    if [ -d ${TEST_DIR}/rootfs/dev ] && [ -f ${TEST_DIR}/rootfs/etc/hostname ]; then
        echo 1
    else
        echo 0
    fi
}

TestMakedevsUsage() {
    makedevs ./rootfs
}

TestMakedevsHelp() {
    makedevs --help
}

TestMakedevsVersion() {
    makedevs --version
}
