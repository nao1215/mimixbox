TestWhichExistBinary() {
    which mimixbox
}

TestWhichNoExistBinary() {
    which no_exist_binary
}

TestWhichThreeBinary() {
    which mimixbox cat tac
}

TestWhichOneOfThreeBinNotExist() {
    which mimixbox not_exist_binary tac
}

TestWhichWithoutOperand() {
    which
}

TestWhichDataFromPipe() {
    echo "test" | which
}