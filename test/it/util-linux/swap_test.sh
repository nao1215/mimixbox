TestSwaponSummary() { swapon -s | sed -n '1p' | grep -c 'Filename'; }
TestSwapoffNoArg() { swapoff 2>/dev/null; echo "rc=$?"; }
