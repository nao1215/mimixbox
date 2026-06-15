# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.39.0] - 2026-06-15

This release grows the applet count to 449 by landing eight issues in parallel:
six BusyBox roadmap batches and two migrations of archived `nao1215` projects.
The batches add their first slices — the prioritized low-risk / high-value
commands are fully implemented and tested, while privileged or kernel-facing
commands are registered as argument-validating, capability-gated applets that
fail deterministically with documented errors (no silent no-ops). The applet
registry is regenerated from the package tree, so every new command appears in
`mimixbox --list` and the README command table.

### Added

- archival (#244): `bzip2` and the `lzop` / `lzopcat` / `unlzop` family with real
  round-trip compression (`-c`/`-d`/`-k`/`-f`/`-t`), plus `dpkg-deb` and `dpkg`
  for read-only local `.deb` inspection (`-c`/`-f`/`-e`/`-I`) and path-safe
  extraction (`-x`/`-X`); package-database operations are rejected with a
  documented error.
- networking client (#246, 28 commands): `ipcalc`; `netcat` (a front over `nc`);
  the shared read-only `ip` / `ipaddr` / `iplink` / `iproute` / `ipneigh` /
  `iprule` family; `ifconfig` / `route` / `netstat` / `arp` inspection;
  `nslookup` / `whois` / `dnsdomainname` over injectable backends; and
  loopback-tested `telnet` / `tftp` / `ftpget` / `ftpput` / `pscan` /
  `ether-wake`. `traceroute*` / `ping6` / `arping` / `tc` / `iptunnel` /
  `nameif` / `slattach` parse and validate, then report capability errors.
- networking daemons (#247, 26 commands): foreground loopback `httpd`,
  `tcpsvd` / `udpsvd`, `inetd`, `fakeidentd`, `dnsd`, `tftpd`, `telnetd`,
  `ftpd`; `dumpleases`; transport-injected `udhcpc` / `udhcpc6` / `udhcpd` /
  `ntpd`; loopback-TLS `ssl_client` / `ssl_server`; and config-driven `ifup` /
  `ifdown`. `brctl` / `ifenslave` / `tunctl` / `vconfig` / `zcip` / `nbd-client` /
  `dhcprelay` / `ifplugd` validate and serialize their plans behind documented
  capability gates.
- console-tools (#252, 13 commands): `bbconfig`, `chat`, and `setserial`, plus
  parse/validate-complete `showkey`, `dumpkmap` / `loadkmap`, `loadfont` /
  `setfont`, `adjtimex`, `microcom`, `rx`, `conspy`, and `openvt`; keymap and
  font binary formats are decoded in a shared, fully unit-tested package.
- embedded (#253, 16 commands): `getfattr` / `setfattr` (real xattr round-trip),
  `lsscsi`, `makedevs`, `volname`, and `readahead`, plus backend-interfaced
  `devmem`, `i2cdetect` / `i2cget` / `i2cset` / `i2cdump`, `partprobe`,
  `raidautorun`, `resume`, `seedrng`, and `watchdog`. Every hardware touchpoint
  is behind an injectable interface so tests need no real device.
- compat (#254, 25 commands): module `lsmod` / `modinfo` / `depmod` (plus gated
  `insmod` / `rmmod` / `modprobe`); SELinux `getenforce` / `selinuxenabled` /
  `sestatus` / `getsebool` / `matchpathcon` (plus gated `setenforce` / `chcon` /
  `runcon` / `restorecon` / `setfiles` / `load_policy` / `setsebool`); mail
  `makemime` / `reformime` / `sendmail` / `popmaildir`; and print `lpr` / `lpq` /
  `lpd` over a temporary spool directory.
- shellutils: `leadtime` (#256), migrated from the archived `nao1215/leadtime`,
  computes GitHub Pull Request lead-time statistics
  (`leadtime stat --owner=OWNER --repo=REPO`: total, max/min/sum/average/median,
  with per-PR detail under `--all`). It supports text, `--json`, and `--markdown`
  output; `--exclude-bot` / `--exclude-pr` / `--exclude-user` filters;
  `--base-url` for GitHub Enterprise and local test servers; and
  `LT_GITHUB_ACCESS_TOKEN` with a `GITHUB_TOKEN` fallback. Read-only REST access
  only.
- textutils: `sqluv` (#255), a script-friendly SQL viewer and query runner over
  local CSV/TSV/LTSV files and SQLite3 databases, migrated from the archived
  `nao1215/sqluv` project. The headless path
  (`sqluv --execute 'SELECT ...' SOURCE --output=table|csv|tsv|json`) is fully
  implemented and CI-tested: it loads each source into an in-memory SQLite
  database (one table per delimited file, plus the tables of any SQLite source),
  runs the SQL once, and prints the result. Transparently compressed inputs
  (`.gz`, `.bz2`, `.xz`, `.zst`) are supported. Database access is read-only by
  default (`--read-only`), query history is written to a configurable file
  (`--history-file`, defaulting to a temp path so it never touches the real home
  directory), and a minimal TUI viewer is started when no `--execute` is given.
  HTTPS, S3, and remote RDBMS DSNs are not migrated yet and fail with
  deterministic, documented errors.

### Fixed

- test: stabilized the new loopback networking tests — drain the client before
  closing the `netcat` echo server to avoid an RST, synchronize the `tftp` put
  test on server completion, and guard the `httpd` test's shared output buffer
  for `-race` — and switched the `dpkg` end-to-end specs to a committed `.deb`
  fixture under `test/it/testdata/`, since the isolated E2E `PATH` resolves
  `tar`/`ar` to the MimixBox applets.

## [0.38.0] - 2026-06-11

This release grows the applet count to 297 by completing three more BusyBox
roadmap batches and starting a fourth. The `[procps]` (#245) and the two
`[util-linux]` batches (#248 device/system and #249 filesystem/device) are now
complete, and the `[loginutils]` batch (#250) is underway. Several filesystem
creators are byte-validated against the host `fsck.minix`, `fsck.vfat`, and
`fsck.ext2`.

### Added

- procps (#245, complete): `top` (batch snapshot), `smemcap`, `klogd`,
  `nmeter`, `logread`, `syslogd` (a minimal datagram-socket logging daemon), and
  `powertop` (one-shot power-supply report).
- util-linux filesystem/device (#249, complete): `flock`, `findfs`, `lsattr`,
  `chattr`, `mount`/`umount` (read-only listing), `mkswap`, `swapon`/`swapoff`,
  `fsfreeze`, `fstrim`, `blkdiscard`, `losetup`, `eject`, `tune2fs`, `unshare`,
  `nsenter`, `freeramdisk`, `pivot_root`, `rtcwake`, `switch_root`, `fdflush`,
  `fdformat`, `fatattr`, `fbset`, `mdev`, `uevent`, `fdisk` (MBR listing), and
  the filesystem creators/checkers `mkfs.minix`/`fsck.minix`,
  `mkfs.vfat`/`mkdosfs`, `mke2fs`/`mkfs.ext2`, `fsck` (type detection), and
  `mkfs.reiser` (documented deprecation notice). The minix, FAT16, and ext2
  creators produce images that pass the host `fsck.*` checkers.
- loginutils (#250, in progress): `nologin`, `run-parts`, `mkpasswd` (crypt
  hashing matching `openssl passwd`), `chpasswd`, `runlevel`, `addgroup`,
  `delgroup`, `adduser`, `deluser`, `crontab`, `start-stop-daemon`, and `crond`
  (foreground cron with a Vixie-style schedule matcher).

### Changed

- Privileged and kernel-facing applets keep their syscalls, ioctls, and
  account/spool databases behind injectable helpers so the whole suite stays
  hermetically unit-tested without root or real hardware.
- Account-mutating applets (`chpasswd`, `addgroup`, `delgroup`, `adduser`,
  `deluser`) replace `/etc/passwd`, `/etc/shadow`, and `/etc/group` atomically
  (temp file plus rename) so an interrupted write cannot corrupt the database.

### Fixed

- Hardened the flaky `nc` loopback integration test: the one-shot listener is
  restarted on a fresh port each attempt, the client send is retried, and only
  the first received line is matched.

## [0.37.0] - 2026-06-09

This release roughly doubles the applet count (to 208) by filling in the
BusyBox/coreutils gaps surfaced in the 2026-06-08 project review. The
`[coreutils]` batch is complete; the `[compat]` shell front-ends are done; and
the `[archival]` and `[util-linux]` batches are well underway.

### Added

- coreutils: `bc` and `dc` (calculators), `ed` (line editor), `ls` (with -a,
  -A, -d, -l, -F, -h, -R), `man` (manual-page lookup with MANPATH and gzip),
  `more`/`less` (pagers), `tree`, `factor`, `tsort`, `nice`, `time`, `fsync`,
  `usleep`, `uuencode`/`uudecode`, the `sum`/`crc32`/`sha384sum`/`sha3sum`
  checksums, the `egrep`/`fgrep` grep wrappers, and `users`/`w`. This completes
  the coreutils roadmap batch.
- compat: the `[`/`[[` test aliases, the `sh`/`ash`/`hush`/`bash` shell
  front-ends over mbsh, the `busybox` multi-call dispatcher, `cttyhack`, and
  `unit`.
- archival: `xz`/`unxz`/`xzcat` and `lzma`/`unlzma`/`lzcat` (via
  github.com/ulikunitz/xz), the `zcat`/`bzcat` decompress-to-stdout aliases, and
  `pipe_progress`.
- util-linux: `hexdump`/`hd`, `getopt`, `setsid`, `fallocate`,
  `script`/`scriptreplay`, `setarch`/`linux32`/`linux64`, `last`, and `renice`.
- `vi` gained the everyday primitives: counts, the `w`/`b`/`e` word motions,
  `yy`/`p`/`P` yank and paste, `u` undo, and `/`/`?`/`n`/`N` search.
- `mbsh` gained quoted tokenization, parameter and environment expansion, and
  `;`/`&&`/`||` separators, pipelines, and redirections.
- `wget` gained `-P`, `-c`, `-T`, `-t`, and `--user-agent`; `cp` gained the
  `-L`/`-P`/`-H`/`-d` symlink-dereference controls.
- Every applet now has a self-describing `--help` (synopsis, options, examples,
  exit status, and notes) aimed at both humans and LLMs.

### Changed

- The applet registry is generated from the package list instead of a
  hand-maintained init, so a new applet is wired up by adding its package.
- The version string is injected from the git tag across `make build` and the
  GoReleaser release, and the release archive ships a self-contained installer.
- The Go directive moved to 1.24.

### Fixed

- `pager` (more/less): use a real tty check so `/dev/null` is not treated as a
  terminal, and drop the per-line length cap.
- `mbsh`: share the stdin position so a launched command reads the remaining
  input, and close already-open pipe fds when `os.Pipe` fails.
- `vi`: decode terminal escape sequences as motions instead of running them as
  commands.
- `wget`: validate the resume `Content-Range` and retry mid-transfer failures.
- `pidof` is now registered, so MimixBox ships and tests its own implementation.

## [0.36.0] - 2026-06-09

### Changed

- The top-level `mimixbox` dispatcher was rewritten without the go-flags
  hacks into a single testable function. `mimixbox --help`/`--version`/
  `--list` now exit 0 and print to stdout; an unknown command/option prints
  the error and the supported-applet list to stderr and exits non-zero.
  Dispatch is decided by the first argument, so an applet's own flags always
  reach it (`mimixbox cp -f a b` runs `cp -f`, no longer mistaken for
  `--full-install`), and an install/remove target may share a basename with
  an applet.
- `cp` is closer to GNU cp: `-R` is a proper alias of `-r`, `-a` means `-rp`,
  and `-n`/`--no-clobber` skips existing destinations.
- `timeout` gained `-k`/`--kill-after`, which sends SIGKILL when the command
  ignores the initial signal.
- `internal/lib`: `Question` loops instead of recursing, and `FromPIPE` uses
  `io.ReadAll` instead of the deprecated `ioutil.ReadAll`.
- Custom-parsed applets now share a single `--help`/`--version` contract via
  `command.HandleHelpVersion`: `find --version` prints the version line
  instead of usage text, and standalone `echo` honors a leading `--help`/
  `--version` like GNU's `/usr/bin/echo` (a later `--help` stays literal, so
  `echo foo --help` is unchanged). `true` and `false` delegate to the same
  helper.

### Fixed

- The unit-test suite is hermetic: `internal/lib` file tests build their
  fixtures with `t.TempDir()` (no `/tmp/mimixbox/ut` dependency, parallel
  safe), and tests that need `chown`, loopback listen sockets or netlink now
  `t.Skip` when those are unavailable instead of failing.
- CI installs shellspec from its canonical `install.sh` URL; the old
  `git.io/shellspec` shortlink was shut down and returned 404, breaking the
  integration-test job.

### Added

- `tail` can follow a growing file: `-f`/`--follow[=name|descriptor]` prints
  appended data, `-F` (and `--retry`) re-opens a rotated/recreated file, and
  `-s`/`--sleep-interval` sets the poll interval. Following honors the context
  so cancellation stops the loop without leaking a goroutine.
- Top-level dispatch tests (`cmd/mimixbox`) and shellspec end-to-end specs
  for previously-untested applets: `printenv`, `printf`, `pwd`, `sleep`,
  `sync`, `uuidgen`, `tr`, `kill`, and a `gzip`/`gunzip` roundtrip.

## [0.35.1] - 2026-06-08

### Fixed

- `cp` now copies files and directories with the **source's** permission
  bits instead of a hardcoded 0644/0755, so the execute bit on scripts is
  kept and private directory trees are not widened. `cp -f` now actually
  removes and replaces a destination that cannot be opened.
- `mv` across filesystems (the `EXDEV` copy+remove fallback) preserves mode
  and timestamps and moves directories recursively, instead of dropping
  metadata and failing on directories.
- `dos2unix`, `unix2dos` and `internal/lib.ListToFile` rewrite files
  atomically (write a temp file, then rename) and preserve the original
  mode, so an interrupted write or full disk no longer truncates the
  original file.
- `watch` runs the child command with the context, so cancelling watch
  (Ctrl-C) interrupts a hung child instead of blocking on it.

### Changed

- `cat` and `nl` stream their input line by line instead of reading whole
  files into memory, so they work in constant space on large files and
  pipes; `internal/textproc.Numberer` streams likewise.
- `internal/lib` file tests build their fixtures with `t.TempDir()` instead
  of a shared `/tmp/mimixbox/ut` tree, so `go test ./...` passes on a clean
  checkout and the suite is parallel-safe.

## [0.35.0] - 2026-06-08

### Added

- New textutils applets on the `internal/command` framework, each with
  GNU-style options, table-driven unit tests and shellspec integration tests:
  `comm` (#159), `paste` (#160), `fold` (#161), `fmt` (#162), `split` (#163),
  `shuf` (#164), `rev` (#165), `cksum` (#166), `strings` (#167), `xxd` (#168),
  `base32` (#169).
- New fileutils applets, each with unit tests and shellspec integration tests:
  `stat` (#170), `truncate` (#171), `readlink` (#172), `link` (#173),
  `unlink` (#174), `shred` (#175), `mountpoint` (#176).
- New shellutils system-information applets, each with unit tests and shellspec
  integration tests: `yes` (#177), `uname` (#178), `arch` (#179), `nproc`
  (#180), `hostname` (#181), `logname` (#182), `tty` (#183).
- New shellutils process/system applets, each with unit tests and shellspec
  integration tests: `nohup` (#184), `timeout` (#185), `watch` (#186),
  `free` (#187), `pidof` (#188), `killall` (#189). The command-running
  applets (`nohup`, `timeout`, `watch`) stop parsing their own options at the
  wrapped command so its flags pass through unchanged.
- New jokeutils applets, each with unit tests and shellspec integration tests:
  `fortune` (#190), `banner` (#191), `cowthink` (#192), `nyancat` (#193),
  `cmatrix` (#194). The animated `nyancat`/`cmatrix` expose pure frame helpers
  for testing and degrade gracefully when no terminal is available.
- New ported applets (clean-room, no GPL source copied): `posixer` (#195)
  reports which POSIX utilities are installed; `pwgen` (#201) generates
  random passwords; `unshadow` (#202) merges passwd/shadow for authorized
  auditing; `pwscore` (#203) rates password strength; `http-status-code`
  (#206) explains HTTP status codes. New `netutils` and `securityutils`
  applet categories were introduced.
- New network/system ported applets, with original MIT/BSD-3 attribution
  preserved where applicable and no source copied verbatim: `nc` (#197,
  netcat), `ping` (#198, raw-socket ICMP), `whris` (#199, domain IP/AS
  lookup), `log-collect` (#200, gather log files), `speaker` (#196,
  TTS via an installed engine).
- New securityutils cracking applets (clean-room, no GPL source copied):
  `pwcrack` (#205) audits crypt(3) hashes against a wordlist, and
  `zip-pwcrack` (#204) recovers a ZipCrypto-encrypted archive's password.
  Hashing uses the permissively licensed `github.com/GehirnInc/crypt`; the
  ZipCrypto cipher is implemented from the documented PKWARE algorithm.
- `internal/auth`: cross-platform basic authentication (#34). The default
  static build verifies passwords against `/etc/shadow` via crypt(3); a PAM
  backend can be selected with `-tags pam` so the no-PAM build never needs
  cgo or libpam.
- A `Docker` CI workflow that builds the image from the local source tree
  and verifies the in-image `mimixbox` binary runs, so building MimixBox in
  Docker stays working (#4).

### Changed

- Raised overall test coverage above 80% (octocov target) by adding
  unit tests for the previously-untested `internal/lib` helpers (string,
  type, signal, option, path, shell, crypto, id, net, shadow, version)
  and `internal/version`.

## [0.34.0] - 2026-06-08

### Added

- Many new applets, each with GNU-style options, unit tests and shellspec
  integration tests: `cal`, `chmod`, `dd`, `df`, `du`, `od` (#144); `install`,
  `mknod` (#146); `resize` (#147); `find`, `grep`, `xargs` (#148); `tar`,
  `gunzip`, `bunzip2`, `zip`, `unzip` (#149); `ar`, `cpio` (#150); `sed`,
  `diff` (#151); `awk`, `patch` (#152); `vi` (#155); `compress`, `uncompress`
  (#156); `rpm2cpio`, `rpm` (#157).
- `mbsh` grew into a minimally usable interactive shell: `cd` to `$HOME` and
  `cd -`, a cwd-aware prompt, comment lines, `$?` expansion, `exit`/`quit`, and
  a fallback that runs MimixBox applets when a command is not on `PATH` (#153).
- `compress`/`uncompress` share a from-scratch Unix LZW (.Z) codec that is
  byte-compatible with the system `compress` and `gzip -d`.
- `rpm`/`rpm2cpio` share an internal RPM parser (lead, headers, gzip/bzip2
  payloads).

- `internal/command`: a small framework every applet can be built on. An applet
  is now a `Command` that receives its I/O streams and arguments as values, so
  it is testable in memory. Flag parsing moves to [spf13/pflag](https://github.com/spf13/pflag)
  via `command.NewFlagSet`, giving GNU-style options (`--long`, clustered
  `-abc`, `--`, interspersed operands) plus standard `--help` / `--version`.
- `internal/textproc`: pure, unit-tested text logic (counting, line numbering,
  reversal, head/tail) shared by the text applets.
- `internal/version`: a single version string, replacing the per-applet
  version constants.
- `internal/hashsum`: shared digest logic backing md5sum/sha1sum/sha256sum/sha512sum.
- Unit tests for every migrated applet and the new packages.

### Changed

- Migrated **all** applets to the new framework with GNU coreutils option
  behavior: every applet now implements `command.Command`, takes its I/O as
  injected streams, parses flags with pflag, and is covered by unit tests.
  Interactive commands (`rm -i`, `cp -i`, `mv -i`, `sddf`, `mbsh`) read from the
  injected input; network commands (`wget`, `ghrdc`) are tested with `httptest`;
  the terminal games (`lifegame`, `sl`) degrade gracefully without a TTY; and
  `halt`/`poweroff`/`reboot` keep the reboot syscall behind a stubbable hook so
  tests never touch the machine.

### Removed

- Man pages and the pandoc-based generation (`scripts/mkManpages.sh`,
  `docs/man/`); use each command's `--help` instead.
- `NOTICE` and the Japanese introduction docs (`docs/introduction/ja/`).

## [0.33.0] - 2021-12-20
### Added
 - lifegame command.
### Changed
 - id command.
   - Enabled to get the execution user ID with the -u option(#82).
## [0.32.1] - 2021-12-19
### Added
 - gzip command.
 - tr command.
 - poweroff command.
 - reboot command.
### Changed
 - halt command.
   - Fix bug(GitHub Issue #33)
## [0.31.1] - 2021-12-17
### Added
 - add-shell command
 - clear command.
 - halt command. However, this version can not shutdown system (halt have the bug).
 - printenv command.
 - pwd command.
 - remove-shell command.
 - reset command.
 - sync command.
 - valid-shell command.
### Changed
 - Project
   - the classification of directories under internal/applets.
 - mimixbox command
   - Fixed a bug that the mimixbox command causes a runtime error. A runtime error occurred when args[0] is an applet name that does not exist and args[1:] contains the applet name.
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