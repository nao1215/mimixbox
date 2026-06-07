TestPathBasename() {
    path -b /home/nao/test.txt
}

TestPathDirname() {
    path -d /home/nao/test.txt
}

TestPathExtension() {
    path -e /home/nao/test.txt
}

TestPathCanonical() {
    path -c /home/nao/../nao/./test.txt
}

TestPathNoOperand() {
    path
}
