TestMinipsHeader() { minips | sed -n '1p' | grep -c 'COMMAND'; }
TestMinipsRows() { minips | grep -cE '^[0-9]+'; }
