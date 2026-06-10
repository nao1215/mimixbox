TestChrtPrint() { chrt -p $$ | grep -c 'scheduling policy'; }
TestChrtRun() { chrt -o 0 -- echo scheduled; }
