# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.31.00] - 2021-12-17
### Added
 - clear command.
 - halt command. However, this version can not shutdown system.
## [0.30.00] - 2021-12-15
### Added
 - chown command.
 - kill command.
 - wget command.
## [0.29.00] - 2021-12-15
### Added
 - uuidgen command.
 - chgrp command.
 - dirname command (with integration tests.)
### Changed
 - mimixbox
   - Fixed the installation order to be in alphabetical order. Previously, the order of installing Applets was random. 
## [0.28.09] - 2021-12-14
### Changed
 - unix2dos / dos2unix command.
   - Changed to print the message being converted.
   - Fixed a bug that the way pipes are handled changes depending on whether there is an option or not.
 - mv command.
   - Change to be able to move multiple file at same time.
   - Changed to continue processing even if it fails while moving multiple files.
   - Fixed the bug that the directory could not be copied due to an error in the copy destination path creation process.
## [0.28.06] - 2021-12-12
### Changed
 - wc command.
   - Fixed the bug that did not count the number of rows of data passed from PIPE correctly.
 - cat / md5sum / sha1sum / sha256sum / sha512sum command.
   - Fixed a bug that the way pipes are handled changes depending on whether there is an option or not.
 - rm command.
   - Changed to receive data from pipe.
   - When removing multiple files, processing continues even if remove fails in the middle.
 - sddf command.
   - Changed to get the file path as much as possible without stopping the process even if an error occurs.
   - Changed to output "." continuously while getting the file path.
   - Changed to show the size of the deleted files.
   - Important files under /dev and /boot and etc. are excluded from deletion.
   - Speeded up checksum calculation with goroutine.
   - Removed named PIPE from checksum calculation. 
     - The checksum calculation for the named PIPE will stop unless there is writing to the named PIPE. It's looks like deadlock. To avoid this problem, exclude named PIPE from target file list.
 - basename command.
   - Print error message if there is no argument.
   - Matched the result with Coreutils when the user specified an empty string.
   - Matched the result with Coreutils when the user specified multiple arguments.
 - mkfifo command.
   - Changed to continue processing even if multiple named pipes fail while creating.
   - Changed to print the path of the file that failed to be created on error.
 - dos2unix command.
   - Changed to print the message being converted.
## [0.28.00] - 2021-12-08
### Added
 - sddf command. Search & Delete Duplicated Files.
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
 - touch command.
   - Continue processing even if an error occurs
 - cp command.
   - Print error if the -r option is not attached when copying the directory.
   - Print error if the copy destination is in the hierarchy below the copy source directory.
   - Fixed the bug that the directory hierarchy of the source path is copied as it is when copying the directory. The correct process is to copy from the base name of the source path.
 - md5sum / sha1sum / sha256sum / sha512sum command.
   - Print the error if the argument is directory or non-existent file.
   - If PIPE and file path are passed at the same time, the PIPE data will be ignored.
 - which command.
   - Change exit-status frmo succes to error if which command can't find binary.
   - Changed the specification that allows only one command to be searched.
 - nl command.
   - Fixed the bug that the cat command does not refer to the argument when pipe is used.
   - Delete unused line feed.
   - Fixed a bug that line numbers do not match when concatenating PIPE data and files.
   - Print the error when specifying the file that does not exist.
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