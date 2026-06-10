TestPsHeader() { ps | sed -n '1p'; }
TestPsHasProcesses() { ps | grep -cE '^ *[0-9]+ '; }
