TestVmstatHeader() { vmstat | sed -n '2p'; }
TestVmstatData() { vmstat | sed -n '3p' | grep -cE '^[ 0-9]+$'; }
