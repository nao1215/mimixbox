# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
## [0.27.16] - 2021-12-02
### Changed
 - mimixbox command.
   - When an error occurs, the name of the applet that caused the error is displayed.
 - wc command.
   - Fixed the bug that the file specified by the argument is not referenced when the wc command is connected by pipe.
   - Display the count result even if the argument is directory.
   - When an error occurs, wc command display its name.
 - cat command. 
   - Fixed the bug that the cat command does not refer to the argument when pipe is used.
   - Fixed the bug that some lines do not have line numbers when the --number option is specified and there is a line feed code in the file-to-file concatenation.
   - Concatenate here documents and files.
 - mkdir command.
   - Output an error instead of a help message when no argument is specified.
## [0.27.10] - 2021-12-01
### Added
 - Add ShellSpec tetsing framework for integration test.
 - ut(Unit Test)／it（Integration Test）target in Makefile.
 - full-install／remove target in Makefile.
### Changed
 - MimixBox command.
   - Display help message when --install, --full-install, --remove are specified and there is no optional argument.
   - Make the error if the directory specified by the user does not exist when executing --install, --full-install, --remove.
   - Fixed the bug that always determines that the applet does not exist when the user specifies the applet with the full path.
 - Makefile.
   - Display accurate coverage by specifying "-coverpkg=./..." in the unit test.
 - Commands that read file (dos2unix, expand, head, tail, unexpand, wc)
   - Fixed the bug that caused Runtime Error when reading the empty file.
 - mkdir command.
   - Create multiple directories with a single command.
     Previously, an error occurred when specifying multiple directories.
 - cp command.
   - Fixed the bug that files cannot be copied when the copy destination is only the directory name.
   - Fixed the bug that the cp command could not copy the directory with complex tree structure.
 - wc command.
   - Unified output format with Coreutils.
   - Fixed the bug that the -L option was not implemented.
 - unix2dos command.
   - Fixed the bug that print incorrect command name. 

## [0.27.1] - 2021-11-29
### Added
 - sl commad.
 - wc command.
### Changed
- expand environment variables in PATH specified by argument
  (mimixbox, cp, ln, mkdir, mkfifo, mv, rm, rmdir, touch, fakemovie, chroot,
   cat, dos2unix, expand, head, tac, tail, unexpand, unix2dos)
- Fixed the bug that was using the wrong installation path.
## [0.25.1] - 2021-11-27
### Added
 - hostid commad(Does not work properly)
 - md5sum command.
 - seq command.
 - sha1sum／sha256sum／sha512sum command
### Changed
- error output destination of MimixBox from STDOUT to STDERR
- all commands to support redirects.
## [0.22.0] - 2021-11-25
### Added
 - dos2unix/unix2dos command.
 - expand/unexpand command.
 - id command: Because GroupIds requires cgo, id command does not work docker environment.
 - groups command.
 - whoami command.
 ### Changed
 - cowsay command to receive data from PIPE.
 - Since the method to display Version(showVeriosn()) was duplicated, it was converted to a library method.
 ### Deleted
 - sh command. It is a command being implemented and is not planned to be POSIX compliant, so it was deleted.
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