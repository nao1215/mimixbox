TestBasenameFilenameWithExt() {
    basename "/home/nao/test.txt"
}

TestBasenameFilenameWithoutExt() {
    basename "/home/nao/test"
}

TestBasenameHiddenFile() {
    basename "/home/nao/.test"
}

TestBasenameEndsWithThrash() {
    basename "/home/nao/"
}

TestBasenameNoOpertand() {
    basename 
}

TestBasenameRoot() {
    basename "/"
}

TestBasenameEmptyString() {
    basename ""
}

TestBasenameWithThreeArg() {
    basename /bin/basename /home/nao /home
}

TestBasenameThreeArgWithMultipleOption() {
    basename -a /bin/basename /home/nao /home
}

TestBasenameThreeArgWithMultipleAndZeroOption() {
    basename -a -z /bin/basename /home/nao /home
}

TestBasenameWithSuffixOption() {
    basename -s .txt /home/nao/test.txt
}