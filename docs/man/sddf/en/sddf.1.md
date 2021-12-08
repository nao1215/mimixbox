% SDDF(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% December 2021

# NAME

serial –  Search & Delete Duplicatetd File

# SYNOPSIS

**sddf** [OPTIONS] PATH

# DESCRIPTION
**sddf** looks for duplicate files under the specified directory and  
creates a list of them (default: duplicated-file.sddf).  
If the list is executed with sddf arguments, sddf delete the file based  
on the contents of the list.

# EXAMPLES
**Search for duplicate files under the current directory**  

    $ sddf .  

**Remove duplicate files**  

    $  sddf duplicated-file.sddf  

# OPTIONS
**-o, **--output**
:   Specify the file name of the duplicate file list.  

**-h**, **--help**
:   Show help message.

**-v**, **--version**
:   Show sddf command version.

# EXIT VALUES
**0**
:   Success

**1**
:   Error when specifying the argument of the sddf command, or error during file operation

# BUGS
See GitHub Issues: https://github.com/nao1215/mimixbox/issues

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.