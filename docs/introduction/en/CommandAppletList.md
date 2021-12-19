# Command (Applet) List
|Command (Applet) Name | Description|
|:--|:--|
| add-shell| Add shell name to /etc/shells |
|    base64| Base64 encode/decode from FILR(or STDIN) to STDOUT|
|  basename| Print basename (PATH without "/") from file path |
|      cat | Concatenate files and print on the standard output|
|		cawsay | Print message with cow's ASCII art|
|   chgrp  | Change the group of each FILE to GROUP|
|   chown  | Change the owner and/or group of each FILE to OWNER and/or GROUP|
|   chroot | Run command or interactive shell with special root directory|
|chsh      | Change login shell|
|   clear  | Clear teminal |
|       cp | Copy file(s) otr Directory(s) |
|  dirname | Print only directory path |
| dos2unix | Change CRLF to LF|
|     echo | Display a line of text|
|   expand | Convert TAB to N space (default:N=8)|
|fakemovie | Adds a video playback button to the image|
|    false | Do nothing. Return unsuccess(1)|
|    ghrdc | GitHub Relase Download Counter|
|   groups | Print the groups to which USERNAME belongs|
|   gzip   | Compress or uncompress FILEs (by default, compress FILES in-place)|
|   halt   | Halt the system|
|    head  | Print the first NUMBER(default=10) lines |
| hostid   | Print hostid (Host Identity Number, hex)|
|       id | Print User ID and Group ID|
|  ischroot| Detect if running in a chroot|
|   kill   | Kill process or send signal to process|
|   lifegame| Life game (Conway's Game of Life)|
|       ln | Create hard link or symbolic link|
|     mbsh | Mimix Box Shell (In development)|
|    md5sum| Calculate or Check md5sum message digest|
|    mkdir | Make directories|
|    mkfifo | Make FIFO (Named pipe)|
|       mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY|
|       nl| Write each FILE to standard output with line numbers added|
|     path | Manipulate filename path|
|    poweroff| Power off the system|
|    printenv| Print environment variable|
|     pwd  | Print Working Directory|
|    reboot| Reboot the system|
|   remove-shell|Remove shell name from /etc/shells|
|     reset| Reset terminal|
|     rm   | Remove file(s) or directory(s)|
|     rmdir   | Remove directory|
|   sddf   | Search & Delete Dupulicated File|
|   seq   | Print the column of numbers|
|   serial | Rename the file to the name with a serial number|
|    sha1sum| Calculate or Check sercure hash 1 algorithm|
|    sha256sum| Calculate or Check sercure hash 256 algorithm|
|    sha512sum| Calculate or Check sercure hash 512 algorithm|
|           sl| Cure your bad habit of mistyping|
|    sleep | Pause for NUMBER seconds(minutes, hours, days)|
|   sync   | Synchronize cached writes to persistent storage|
|     tac  | Print the file contents from the end to the beginning|
|     tail |  Print the last NUMBER(default=10) lines|
|    touch | Update the access and modification times of each FILE to the current time|
|    tr    | Translate or delete characters|
|     true | Do nothing. Return success(0)|
|  unexpand| Convert N space to TAB (default:N=8)|
|  unix2dos| Change LF to CRLF|
|   uuidgeb| Print UUID (Universal Unique IDentifier|
|  valid-shell| Verify if /etc/shells is valid|
|    wc    |    Word Counter|
|    wget  | The non-interactive network downloader|
|    which | Returns the file path which would be executed in the current environment|
|   whoami | Print login user name|

If you want to see the list of supported commands on the terminal, use the --list option.

```
$ ./mimixbox --list
   base64 - Base64 encode/decode from FILR(or STDIN) to STDOUT
 basename - Print basename (PATH without"/") from file path
      cat - Concatenate files and print on the standard output
    chgrp - Change the group of each FILE to GROUP
    chown - Change the owner and/or group of each FILE to OWNER and/or GROUP
   chroot - Run command or interactive shell with special root directory
    clear - Clear terminal
   cowsay - Print message with cow's ASCII art
       cp - Copy file(s) otr Directory(s)
  dirname - Print only directory path
 dos2unix - Change CRLF to LF
     echo - Display a line of text
   expand - Convert TAB to N space (default:N=8)
fakemovie - Adds a video playback button to the image
    false - Do nothing. Return unsuccess(1)
    ghrdc - GitHub Relase Download Counter
   groups - Print the groups to which USERNAME belongs
     halt - Halt the system
     head - Print the first NUMBER(default=10) lines
   hostid - Print hostid (Host Identity Number, hex)!!!Does not work properly!!!
       id - Print User ID and Group ID
 ischroot - Detect if running in a chroot
     kill - Kill process or send signal to process
       ln - Create hard or symbolic link
     mbsh - Mimix Box Shell
   md5sum - Calculate or Check md5sum message digest
    mkdir - Make directories
   mkfifo - Make FIFO (named pipe)
       mv - Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY
       nl - Write each FILE to standard output with line numbers added
     path - Manipulate filename path
    reset - Reset terminal
       rm - Remove file(s) or directory(s)
    rmdir - Remove directory
     sddf - Search & Delete Duplicated File
      seq - Print a column of numbers
   serial - Rename the file to the name with a serial number
  sha1sum - alculate or Check sercure hash 1 algorithm
sha256sum - alculate or Check sercure hash 256 algorithm
sha512sum - alculate or Check sercure hash 512 algorithm
       sl - Cure your bad habit of mistyping
    sleep - Pause for NUMBER seconds(minutes, hours, days)
     sync - Synchronize cached writes to persistent storage
      tac - Print the file contents from the end to the beginning
     tail - Print the last NUMBER(default=10) lines
    touch - Update the access and modification times of each FILE to the current time
     true - Do nothing. Return success(0)
 unexpand - Convert N space to TAB(default:N=8)
 unix2dos - Change LF to CRLF
  uuidgen - Print UUID (Universal Unique IDentifier
       wc - Word Count
     wget - The non-interactive network downloader
    which - Returns the file path which would be executed in the current environment
   whoami - Print login user name
```