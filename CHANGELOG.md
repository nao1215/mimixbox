# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [0.20.0] - 2021-11-24
### Added
 - dos2unix/unix2dos command.
 - expand/unexpand command.
 - whoami command.
 ### Changed
 - cowsay command to receive data from PIPE.
## [0.16.1] - 2021-11-23
### Added
 - cowsay command. The process of enclosing the message in a frame is incomplete.
 - head command.
 - tail command.
 - ln command.
 - docker target to Makefile. This target was created to test Mimixbox inside Docker.
 ### Changed
 - nl/tac/cat command to receive data from PIPE.
 - Fixed the buffer overflow for the head / tail command
 - Fixed the bug in the cat command. This bug occurs when a standard input is accepted more than once and then an empty enter is received on the next input. In the correct behavior, it is correct to output a blank line, but since the previous input value has been saved, the previous input value is output.
## [0.12.1] - 2021-11-20
### Added
 - base64 command.
 - cp command.
 - nl command.
 - -n option for cat command
### Changed
 - Reduce mimixbox binary size by compile option (7.5MB --> 5.4MB)
 - cat/tac command to receive data from standard input when the argument is "-".
## [0.9.1] - 2021-11-19
### Added
 - basename command.
 - sleep command.
 - tac command.
### Changed
 - The library that was open to the public(pkg) to the internal library.
## [0.6.0] - 2021-11-18
### Added
 - mkfifo command.
 - rm command.
 - rmdir command.
### Changed
 - cat/tac commands does not output a help message even if there is no argument.
### Fixed
 - Bug that misjudged applet arguments as mimixbox arguments.
## [0.3.0] - 2021-11-17
### Added
 - touch command.
 - licenses target to Makefile.

## [0.2.0] - 2021-11-17
### Added
 - mv command.
## [0.1.1] - 2021-11-16
### Added
 - ischroot command.
 - mkJailForDebianFamily.sh that make rootfs for testing chroot/ischroot command.
 - jail target to Makefile. Only work for Debian-based distribution.
### Changed
 - All applet returns error and exit code for main code.

## [0.0.1] - 2021-11-14
### Added
 - mimixbox project.