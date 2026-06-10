TestSysctlRead() { sysctl kernel.ostype; }
TestSysctlAll() { sysctl -a 2>/dev/null | grep -c ' = '; }
