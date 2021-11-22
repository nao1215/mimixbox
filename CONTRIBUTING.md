# Contributing to MimixBox
First of all, thank you for taking the time to contribute.  
The following provides you with some guidance on how to contribute to this project.   

# Pull Requests
MimixBox will make about 100 commands by the time it reaches Version 1.0.0.  
If you find a command that doesn't exist in MimixBox despite the general Unix command, feel free to make a Pull Request.  
It doesn't matter if there are few options. The test is also minimal. I'll fix some bugs later.  
First, increase the number of commands. The quality will be improved to Version 1.0.0 or later.  

I also accept original joke commands and games.
However, please be careful about the license. Only licenses compatible with Apache License version 2.0 will be merged.  
For example, GPLv2 or latter is not merged. Tetris can't be merged either (sorry).

# Code tree
```
project root dir
├── cmd
│   └── mimixbox
│          └── main.go   ※ If you add command, increment the minor version
├── docs
│    └── introduction
│         └──en
│             └── CommandAppletList.md ※ If you add command, add a command description
└── internal
    └── applets
        ├── applet.go ※ applet.go has command execution entry point. See init().
        ├── fileutils
        ├── games
        ├── jokeutils
        ├── shellutils
        └── textutils
```

For example, fileutils directory structure is below.  
One command is managed in one directory. You can increase the number of files in a certain command directory.
```
internal/applets/fileutils/
├── cp
│   └── cp.go
├── ln
│   └── ln.go
├── mkdir
│   └── mkdir.go
├── mkfifo
│   └── mkfifo.go
├── mv
│   └── mv.go
├── rm
│   └── rm.go
├── rmdir
│   └── rmdir.go
└── touch
    └── touch.go
```

If you want to implement a generic method, implement it in the internal library.  
Treat these libraries as mb libraries.  
```
project root directory
 └── internal
      ├── applets
      └── lib
```
