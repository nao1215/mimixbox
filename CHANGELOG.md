# Changelog
All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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