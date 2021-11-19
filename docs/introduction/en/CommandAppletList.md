# Command (Applet) List
|Command (Applet) Name | Description|
|:--|:--|
|    base64| Base64 encode/decode from FILR(or STDIN) to STDOUT|
|  basename| Print basename (PATH without "/") from file path |
|      cat | Concatenate files and print on the standard output|
|   chroot | Run command or interactive shell with special root directory|
|       cp | Copy file(s) otr Directory(s) |
|     echo | Display a line of text|
|fakemovie | Adds a video playback button to the image|
|    false | Do nothing. Return unsuccess(1)|
|    ghrdc | GitHub Relase Download Counter|
|  ischroot| Detect if running in a chroot|
|     mbsh | Mimix Box Shell (In development)|
|    mkdir | Make directories|
|    mkfifo | Make FIFO (Named pipe)|
|       mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY|
|     path | Manipulate filename path|
|     rm   | Remove file(s) or directory(s)|
|     rmdir   | Remove directory|
|   serial | Rename the file to the name with a serial number|
|       sh | Mimix Box Shell (In development)|
|    sleep | Pause for NUMBER seconds(minutes, hours, days)|
|     tac  | Print the file contents from the end to the beginning|
|    touch | Update the access and modification times of each FILE to the current time|
|     true | Do nothing. Return success(0)|
|    which | Returns the file path which would be executed in the current environment|

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