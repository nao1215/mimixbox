# MimixBox - mimic BusyBox on Linux
MimixBox has many Unix commands in the single binary like BusyBox. However, mimixbox aim for the different uses from BusyBox. Specifically, it is supposed to be used in the desktop environment, not the embedded environment.  
Also, the mimixbox project maintainer plan to have a wide range of built-in commands (applets) from basic command provided by Coreutils and others to experimental commands.

# Installation.
The source code and binaries are distributed on the Release Page in ZIP format or tar.gz format. Choose the binary that suits your OS and CPU architecture.
For example, in the case of Linux (amd64), you can install the commands on your system with the following command:

```
$ tar xf mimixbox-0.0.1-linux-arm64.tar.gz
$ cd mimixbox-0.0.1-linux-arm64
$ sudo ./installer.sh
```

# Command (Applet) List
|Command (Applet) Name | Description|
|:--|:--|
|      cat | Concatenate files and print on the standard output|
|   chroot | Run command or interactive shell with special root directory|
|     echo | Display a line of text|
|fakemovie | Adds a video playback button to the image|
|    false | Do nothing. Return unsuccess(1)|
|    ghrdc | GitHub Relase Download Counter|
|  ischroot| Detect if running in a chroot|
|     mbsh | Mimix Box Shell (In development)|
|    mkdir | Make directories|
|       mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY|
|     path | Manipulate filename path|
|   serial | Rename the file to the name with a serial number|
|       sh | Mimix Box Shell (In development)|
|     true | Do nothing. Return success(0)|
|    which | Returns the file path which would be executed in the current environment.|

# Development (How to build)
If you want to contribute to the mimixbox project, get the source code with the following command.
```
$ git clone https://github.com/nao1215/mimixbox.git
$ cd mimixbox
$ make
```

The table below shows the tools used when developing the commands in the mimixbox project.
| Tool | description |
|:-----|:------|
| dep   | Used for managing dependencies for Go projects|
| gobump   | Used for command version control |
| go-licenses | Used for license management of dependent libraries|
| pandoc   | Convert markdown files to manpages |
| make   | Used for build, run, test, etc |
| gzip   | Used for compress man pages |
| install   | Used for install serial binary and document in the system |

# Contact
If you would like to send comments such as "find a bug" or "request for additional features" to the developer, please use one of the following contacts.

- [GitHub Issue](https://github.com/nao1215/mimixbox/issues)
- [mail@Naohiro CHIKAMATSU](n.chika156@gmail.com)
- [Twitter@ARC_AED](https://twitter.com/ARC_AED)

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.  
See [LICENSE](./LICENSE) and [NOTICE](./NOTICE)