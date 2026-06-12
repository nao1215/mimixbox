TestCryptpwStdin() { echo secret | cryptpw -S abcdefgh; }
TestCryptpwMd5() { echo secret | cryptpw -m md5 -S abcdefgh | head -c 12; echo; }
