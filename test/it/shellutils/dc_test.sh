TestDcDivide() { echo '6 3 / p' | dc; }
TestDcScale() { echo '2k 7 3 / p' | dc; }
TestDcExpr() { dc -e '2 10 ^ p'; }
TestDcRegisters() { echo '5 sa 3 la + p' | dc; }
