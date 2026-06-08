# VHS tape for the MimixBox demo GIF. Regenerate with:
#   make build && PATH="$PWD:$PATH" vhs assets/demo.tape
Output "assets/demo.gif"
Set Shell bash
Set FontSize 18
Set Width 1100
Set Height 650
Set Padding 18
Set Theme "Dracula"

Sleep 500ms
Type "# MimixBox packs many Unix commands into one binary"
Enter
Sleep 1.2s

Type "mimixbox --list | head -n 4"
Enter
Sleep 1.8s

Type "mimixbox echo 'hello from mimixbox'"
Enter
Sleep 1.2s

Type "mimixbox seq 1 5 | mimixbox tac"
Enter
Sleep 1.5s

Type "printf 'apple\\nbanana\\ncherry\\n' | mimixbox grep an"
Enter
Sleep 1.5s

Type "mimixbox cal 12 2024"
Enter
Sleep 2s
