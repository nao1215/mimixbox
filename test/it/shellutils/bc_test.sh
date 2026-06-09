TestBcPrecedence() { echo '2 + 3 * 4' | bc; }
TestBcScale() { echo 'scale=2; 7/3' | bc; }
TestBcVars() { printf 'x = 5\nx * x\n' | bc; }
TestBcPower() { echo '2^10' | bc; }
