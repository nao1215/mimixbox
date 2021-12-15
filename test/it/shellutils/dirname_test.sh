TestDirnameAbsFilePath() {
    dirname "/home/nao/test.txt"
}

TestDirnameFilenameWithoutExt() {
    dirname "/home/nao/test"
}

TestDirnameHiddenFile() {
    dirname "/home/nao/.test"
}

TestDirnameEndsWithThrash() {
    dirname "/home/nao/"
}

TestDirnameNoOpertand() {
    dirname 
}

TestDirnameRoot() {
    dirname "/"
}

TestDirnameEmptyString() {
    dirname ""
}

TestDirnameWithThreeArg() {
    dirname /bin/dirname /home/nao /home
}

TestDirnameThreeArgWithZeroOption() {
    dirname -z /bin/dirname /home/nao /home
}

TestDirnameFilenameWithEnvVar() {
    export TEST_DIR="/aaa/bbb/ccc"
    dirname $TEST_DIR/ddd.txt
}
