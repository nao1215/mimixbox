TestLinux32Uname() { linux32 uname -m; }
TestLinux64Uname() { linux64 uname -m; }
TestLinux32Passthrough() { linux32 echo passed; }
TestSetarchArch() { setarch i686 uname -m; }
