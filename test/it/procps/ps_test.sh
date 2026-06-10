TestPsHeader() { ps | sed -n '1p'; }
TestPsHasInit() { ps | grep -c ' init'; }
