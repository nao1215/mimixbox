# Command (Applet) List
|Command (Applet) Name | Description|
|:--|:--|
|    base64| Base64 encode/decode from FILR(or STDIN) to STDOUT|
|  basename| Print basename (PATH without "/") from file path |
|      cat | Concatenate files and print on the standard output|
|		cawsay | Print message with cow's ASCII art|
|   chroot | Run command or interactive shell with special root directory|
|       cp | Copy file(s) otr Directory(s) |
| dos2unix | Change CRLF to LF|
|     echo | Display a line of text|
|   expand | Convert TAB to N space (default:N=8)|
|fakemovie | Adds a video playback button to the image|
|    false | Do nothing. Return unsuccess(1)|
|    ghrdc | GitHub Relase Download Counter|
|   groups | Print the groups to which USERNAME belongs|
|    head  | Print the first NUMBER(default=10) lines |
| hostid   | Print hostid (Host Identity Number, hex)|
|       id | Print User ID and Group ID|
|  ischroot| Detect if running in a chroot|
|       ln | Create hard link or symbolic link|
|     mbsh | Mimix Box Shell (In development)|
|    md5sum| Calculate or Check md5sum message digest|
|    mkdir | Make directories|
|    mkfifo | Make FIFO (Named pipe)|
|       mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY|
|       nl| Write each FILE to standard output with line numbers added|
|     path | Manipulate filename path|
|     rm   | Remove file(s) or directory(s)|
|     rmdir   | Remove directory|
|   seq   | Print the column of numbers|
|   serial | Rename the file to the name with a serial number|
|    sha1sum| Calculate or Check sercure hash 1 algorithm|
|    sha256sum| Calculate or Check sercure hash 256 algorithm|
|    sha512sum| Calculate or Check sercure hash 512 algorithm|
|           sl| Cure your bad habit of mistyping|
|    sleep | Pause for NUMBER seconds(minutes, hours, days)|
|     tac  | Print the file contents from the end to the beginning|
|     tail |  Print the last NUMBER(default=10) lines|
|    touch | Update the access and modification times of each FILE to the current time|
|     true | Do nothing. Return success(0)|
|  unexpand| Convert N space to TAB (default:N=8)|
|  unix2dos| Change LF to CRLF|
|    wc    |    Word Counter|
|    which | Returns the file path which would be executed in the current environment|
|   whoami | Print login user name|

If you want to see the list of supported commands on the terminal, use the --list option.

```
$ mimixbox --list
      cat - Concatenate files and print on the standard output
   chroot - Run command or interactive shell with special root directory
     echo - Display a line of text
fakemovie - Adds a video playback button to the image
    false - Do nothing. Return unsuccess(1)
    ghrdc - GitHub Relase Download Counter
 ischroot - Detect if running in a chroot
     mbsh - Mimix Box Shell
    mkdir - Make directories
   mkfifo - Make FIFO (named pipe)
       mv - Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY
     path - Manipulate filename path
       rm - Remove file(s) or directory(s)
    rmdir - Remove directory
   serial - Rename the file to the name with a serial number
       sh - Mimix Box Shell
    touch - Update the access and modification times of each FILE to the current time
     true - Do nothing. Return success(0)
    which - Returns the file path which would be executed in the current environment
```