TestLsofSelf() { lsof -p $$ | grep -c 'cwd'; }
TestLsofHeader() { lsof -p $$ | sed -n '1p'; }
