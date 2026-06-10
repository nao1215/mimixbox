TestNmeterLiteral() { nmeter 'hello %% world'; }
TestNmeterMem() { nmeter 'mem:%M' | grep -cE 'mem:[0-9]+M'; }
