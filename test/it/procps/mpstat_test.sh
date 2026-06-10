TestMpstatHeader() { mpstat | sed -n '1p'; }
TestMpstatAll() { mpstat | grep -c '^all '; }
