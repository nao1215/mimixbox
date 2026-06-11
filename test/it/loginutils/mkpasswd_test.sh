TestMkpasswdSha512() { mkpasswd -m sha-512 -S abcdefgh secret; }
TestMkpasswdStdin() { echo "frompipe" | mkpasswd -m md5 -S abcdefgh | head -c 12; echo; }
