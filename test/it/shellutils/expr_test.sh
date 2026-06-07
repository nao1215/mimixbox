TestExprAdd() {
    expr 6 + 7
}

TestExprMul() {
    expr 3 \* 4
}

TestExprGroup() {
    expr \( 1 + 2 \) \* 3
}

TestExprLength() {
    expr length abcd
}

TestExprFalse() {
    expr 0
}
