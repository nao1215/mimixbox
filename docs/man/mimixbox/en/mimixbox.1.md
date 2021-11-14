% MIMIXBOX(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% November 2021

# 名前

MimixBox – mimic BusyBox which has many Unix commands (applet) in the single binary.

# SYNOPSIS

**mimixbox** [applet [arguments]...] [OPTIONS]

# DESCRIPTION
**mimixbox** has many Unix commands in the single binary like BusyBox. However,  
mimixbox aim for the different uses from BusyBox. Specifically, it is supposed  
to be used in the desktop environment, not the embedded environment. Also, the  
mimixbox project maintainer plan to have a wide range of built-in commands  
(applets) from basic command provided by Coreutils and others to experimental  
commands.

# Command（applet）list
**Common unix commands（applets）**  
cat, chroot, echo, false, mkdir, path, serial, sh, true, which

**MimixBox Original commands（applets）**  
fakemovie, ghrdc, mbsh, path, serial

# OPTIONS
**-i**, **--install**
:   Create symbolic links for commands that don't exist on the system.

**-f**, **--full-install**
:   Create symbolic links regardless of system state.

**-h**, **--help**
:   Show this help message.

**-l**, **--list**
:   Show command name provided by mimixbox.

**-r**, **--remove**
:   Remove symbolic links for commands provided by mimixbox.

**-v**, **--version**
:   Show mimixbox command version.

# EXIT VALUES
**0**
:   Success

**1**
:   If you specify a non-existent applet name, or if the option is invalid,
    or if an error occurs in applet

**2**
:   Errors that occur with some applets (eg ischroot, etc.)

# BUGS
See GitHub Issues: https://github.com/nao1215/mimixbox/issues

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.