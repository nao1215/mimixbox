% GHRDC(1)
% Naohiro CHIKAMATSU <n.chika156@gmail.com>
% September 2021

# NAME

ghrdc –  shows the number of release file downloads in the repository using GitHub API.

# SYNOPSIS

**ghrdc** [OPTIONS] USER_NAME/RPOSITORY_NAME

# DESCRIPTION
**ghrdc** shows the number of Release file downloads in the repository.  
By default, it shows the number of downloads for the latest release.  
Because ghrdc does not authenticate with the GitHub API, it has the following restrictions.   
- It can only be run 60 times per hour.  
- Unable to get information in Organization repository.  

# EXAMPLES
**Get the number of downloads for the latest release**  
    $ ghrdc nao1215/serial  
      [Name(Version)]             :Version1.0.2: Release files with installer scripts.  
      [Release Date]              :2020-11-23 05:28:11 +0000 UTC  
      [Binary Download Count]     :177  
      [Source Code Download Count]:0  

# OPTIONS
**-a**, **--all**
:   Show total number of downloads per release.

**-t**, **--total**
:   Show total number of downloads for all releases.

**-h**, **--help**
:   Show help message.

**-v**, **--version**
:   Show ghrdc command version.


# EXIT VALUES
**0**
:   Success

**1**
:   There is an error in the argument of the path command or GitHub API runtime error.

# BUGS
See GitHub Issues: https://github.com/nao1215/mimixbox/issues

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.