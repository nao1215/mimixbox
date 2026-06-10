TestIonicePrint() { ionice -p $$ | grep -cE 'prio|idle'; }
TestIoniceRun() { ionice -c 3 -- echo idled; }
