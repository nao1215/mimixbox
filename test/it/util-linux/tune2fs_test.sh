TestTune2fsNotExt() {
    printf 'not a filesystem' > "$TEST_DIR/bad.img"
    tune2fs -l "$TEST_DIR/bad.img" 2>/dev/null
    echo "rc=$?"
}
TestTune2fsHelp() { tune2fs --help; }
