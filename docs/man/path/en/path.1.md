% PATH(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% September 2021

# NAME

path – command for manipulating filename path.

# SYNOPSIS

**path** [OPTIONS] DIRECTORY_PATH

# DESCRIPTION
**path**  extracts the directory name, file name, and extension  
from the path. It also create canonical path and absolute path.

# EXAMPLES
**Get absolute path**  
    $ path -a path  
      /home/nao/.go/src/otter/path

**Get basename**  
    $ path -b /etc/systemd/pstore.conf   
      pstore.conf

**Get canonical path**  
    $ path -c cmd/../scripts/installer.sh   
      scripts/installer.sh

**Get dirctory name**  
    $ path -d /etc/ssh/ssh_config  
      /etc/ssh

**Get file extension**  
    $ path -e go.mod  
      .mod

# OPTIONS
**-a**, **--absolute**
:   Print absolute path.

**-b**, **--basename**
:   Print basename (filename).

**-c**, **--canonical**
:    Print canonical path (default).

**-d**, **--dirname**
:   Print path without filename.

**-e**, **--extension**
:   Print file extention.

**-h**, **--help**
:   Show help message.

**-v**, **--version**
:   Show path command version.

# EXIT VALUES
**0**
:   Success

**1**
:   There is an error in the argument of the path command.

# BUGS
See GitHub Issues: https://github.com/nao1215/mimixbox/issues

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.