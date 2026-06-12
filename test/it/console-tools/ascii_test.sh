TestAsciiLineCount() { ascii | grep -c .; }
# Code 65 (0x41) must map to the letter A on the same line.
TestAsciiCapitalA() { ascii | grep '0x41' | grep -c 'A'; }
