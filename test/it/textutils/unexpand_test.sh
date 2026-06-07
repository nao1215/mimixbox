TestUnexpandPipe() {
    printf '        a\n' | unexpand
}

TestUnexpandAll() {
    printf 'a        b\n' | unexpand -a
}

TestUnexpandNoExistFile() {
    unexpand /no_exist_file
}
