<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/mimixbox/coverage.svg)
[![Build](https://github.com/nao1215/mimixbox/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nao1215/mimixbox/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml/badge.svg?branch=main&event=push)](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml)
[![IntegrationTest](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml/badge.svg?event=push)](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml)
![GitHub](https://img.shields.io/github/license/nao1215/mimixbox)
![GitHub all releases](https://img.shields.io/github/downloads/nao1215/mimixbox/total)

![mimixbox-logo](./doc/image/mimixbox-logo-truss.jpg)

MimixBox packs many Unix commands into a single binary, like BusyBox. Unlike BusyBox, it targets the desktop environment rather than embedded systems. The project aims for a wide range of built-in commands (applets), from the basics provided by Coreutils to its own experimental commands.

![demo](./assets/demo.gif)

## Commands

The list below is generated from the registered applets by `make command-list`, so it never drifts from the binary. You can also run `mimixbox --list` to print it on the terminal.

<!-- COMMAND_LIST_START -->
There are 280 commands. Run `mimixbox --list` to see them on the terminal.

| Command | Description |
|:--|:--|
| [ | Evaluate a conditional expression (test alias requiring ]) |
| [[ | Evaluate a conditional expression (test alias requiring ]]) |
| add-shell | Add shell name to /etc/shells |
| ar | Create, modify and extract from archives |
| arch | Print machine hardware name (same as uname -m) |
| ash | Command interpreter (MimixBox mbsh compatibility front-end) |
| awk | Pattern scanning and processing language |
| banner | Print a string as large ASCII-art letters |
| base32 | Base32 encode/decode from FILE(or STDIN) to STDOUT |
| base64 | Base64 encode/decode from FILE(or STDIN) to STDOUT |
| basename | Print basename (PATH without "/") from file path |
| bash | Command interpreter (MimixBox mbsh compatibility front-end) |
| bc | An arbitrary-precision calculator language |
| blkdiscard | Discard sectors on a block device |
| blkid | Identify the filesystem type of a device or image |
| blockdev | Report block device properties |
| bunzip2 | Decompress bzip2 (.bz2) files |
| busybox | BusyBox-style multi-call front-end for MimixBox applets |
| bzcat | Decompress bz2 data to standard output |
| cal | Display a calendar |
| cat | Concatenate files and print on the standard output |
| chattr | Change ext2/ext4 file attributes |
| chgrp | Change the group of each FILE to GROUP |
| chmod | Change file mode bits |
| chown | Change the owner and/or group of each FILE to OWNER and/or GROUP |
| chroot | Run command or interactive shell with special root directory |
| chrt | Get or set a process's real-time scheduling attributes |
| cksum | Print CRC checksum and byte count of each file |
| clear | Clear terminal |
| cmatrix | Show the falling-glyph digital rain effect |
| cmp | Compare two files byte by byte |
| comm | Compare two sorted files line by line |
| compress | Compress files with LZW (.Z) |
| cowsay | Print message with cow's ASCII art |
| cowthink | Print message in a cow's thought bubble |
| cp | Copy file(s) to Directory(s) |
| cpio | Copy files to and from archives |
| crc32 | Print the CRC-32 checksum of each file |
| cttyhack | Run PROGRAM with the current stdio (no controlling-TTY trick) |
| cut | Remove sections from each line of files |
| date | Print or set the system date and time |
| dc | Reverse-Polish (stack) desk calculator |
| dd | Convert and copy a file |
| df | Report file system disk space usage |
| diff | Compare files line by line |
| dirname | Print only directory path |
| dmesg | Print or control the kernel ring buffer |
| dos2unix | Change CRLF to LF |
| du | Estimate file space usage |
| echo | Display a line of text |
| ed | A line-oriented text editor |
| egrep | Search with extended regular expressions (grep -E) |
| eject | Eject removable media |
| env | Run a program in a modified environment / print the environment |
| expand | Convert TAB to N space (default:N=8) |
| expr | Evaluate expressions |
| factor | Print the prime factors of each NUMBER |
| fakemovie | Adds a video playback button to the image |
| fallocate | Preallocate or extend space for a file |
| false | Do nothing. Return failure(1) |
| fatattr | Show or change FAT file attributes |
| fbset | Show the framebuffer video mode |
| fdflush | Flush a floppy device's buffers |
| fdformat | Low-level format a floppy device |
| fdisk | List the MBR partition table |
| fgrep | Search for fixed strings (grep -F) |
| find | Search for files in a directory hierarchy |
| findfs | Find a filesystem by label or UUID |
| flock | Run a command under an advisory file lock |
| fmt | Simple optimal text formatter |
| fold | Wrap each input line to fit in specified width |
| fortune | Print a random, hopefully interesting, adage |
| free | Display amount of free and used memory in the system |
| freeramdisk | Free the memory used by a ramdisk |
| fsck | Detect and report a filesystem type |
| fsck.minix | Check a Minix filesystem |
| fsfreeze | Suspend or resume a filesystem |
| fstrim | Discard unused blocks on a filesystem |
| fsync | Flush a file's data to storage with fsync(2) |
| fuser | Identify processes using a file |
| getopt | Parse command options (enhanced, like util-linux getopt) |
| ghrdc | GitHub Release Download Counter |
| grep | Print lines that match patterns |
| groups | Print the groups to which USERNAME belongs |
| gunzip | Decompress gzip (.gz) files |
| gzip | Compress or uncompress FILEs (by default, compress FILES in-place) |
| halt | Halt the system |
| hd | Dump a file in canonical hex+ASCII (hexdump -C) |
| head | Print the first NUMBER(default=10) lines |
| hexdump | Display a file in hexadecimal (and other formats) |
| hostid | Print the numeric identifier (in hexadecimal) for the current host |
| hostname | Show the system's host name |
| http-status-code | Explain HTTP status codes and their RFC references |
| hush | Command interpreter (MimixBox mbsh compatibility front-end) |
| hwclock | Read the hardware (RTC) clock |
| id | Print User ID and Group ID |
| install | Copy files and set attributes |
| ionice | Get or set process I/O scheduling class and priority |
| iostat | Report CPU and device I/O statistics |
| ipcrm | Remove System V IPC objects by id |
| ipcs | Show System V IPC facilities status |
| ischroot | Detect if running in a chroot |
| kill | Kill process or send signal to process |
| killall | Kill processes by name |
| killall5 | Send a signal to all processes |
| klogd | Forward kernel messages to the system log |
| last | Show a listing of last logged-in users |
| less | Page through text one screen at a time |
| lifegame | Life game (Conway's Game of Life) |
| link | Create a hard link to a file |
| linux32 | Run a program with a 32-bit execution domain |
| linux64 | Run a program with a 64-bit execution domain |
| ln | Create hard or symbolic link |
| log-collect | Gather system log files into one directory |
| logger | Write a message to the system log |
| logname | Print the name of the current user |
| logread | Show the system log |
| losetup | List active loop devices |
| ls | List directory contents |
| lsattr | List ext2/ext4 file attributes |
| lsblk | List information about block devices |
| lsof | List open files of processes |
| lspci | List PCI devices |
| lsusb | List USB devices |
| lzcat | Decompress lzma data to standard output |
| lzma | Compress or decompress files (lzma) |
| man | Display a manual page |
| mbsh | Mimix Box Shell |
| md5sum | Calculate or Check md5sum message digest |
| mesg | Display or control write access to your terminal |
| minips | Minimal process lister (PID, user, command) |
| mkdir | Make directories |
| mkdosfs | Create a FAT16 filesystem |
| mkfifo | Make FIFO (named pipe) |
| mkfs.minix | Create a Minix filesystem |
| mkfs.vfat | Create a FAT16 filesystem |
| mknod | Make block or character special files |
| mkswap | Set up a Linux swap area |
| mktemp | Create a temporary file or directory |
| more | Page through text one screen at a time |
| mount | List the mounted filesystems |
| mountpoint | See if a directory is a mountpoint |
| mpstat | Report per-processor CPU statistics |
| mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY |
| nc | Read and write data across network connections |
| nice | Run a command with an adjusted niceness |
| nl | Write each FILE to standard output with line numbers added |
| nmeter | Print system statistics from a format string |
| nohup | Run a command immune to hangups, with output to a non-tty |
| nproc | Print the number of processing units available |
| nsenter | Run a program in another process's namespaces |
| nyancat | Animate the rainbow-trailing Nyan Cat |
| od | Dump files in octal and other formats |
| paste | Merge lines of files |
| patch | Apply a diff file to an original |
| path | Manipulate filename path |
| pgrep | Find process IDs by name |
| pidof | Find the process ID of a running program |
| ping | Send ICMP ECHO_REQUEST to network hosts |
| pipe_progress | Copy stdin to stdout, printing progress dots to stderr |
| pivot_root | Change the root filesystem |
| pkill | Signal processes by name |
| pmap | Report the memory map of a process |
| posixer | Report which POSIX utilities are installed |
| poweroff | Power off the system |
| powertop | Report the system power supplies |
| printenv | Print environment variable |
| printf | Format and print data |
| ps | Report a snapshot of the current processes |
| pstree | Display the process tree |
| pwcrack | Audit crypt(3) password hashes against a wordlist |
| pwd | Print Working Directory |
| pwdx | Print the working directory of a process |
| pwgen | Generate random passwords for authorized testing |
| pwscore | Estimate the strength of a password |
| rdate | Get the time from a remote host (RFC 868) |
| rdev | Print the root filesystem device |
| readlink | Print resolved symbolic links or canonical file names |
| readprofile | Summarize the kernel profiling buffer |
| realpath | Print the resolved absolute file name |
| reboot | Reboot the system |
| remove-shell | Remove shell name from /etc/shells |
| renice | Alter the priority of running processes |
| reset | Reset terminal |
| resize | Print commands to set the terminal size |
| rev | Reverse the order of characters in every line |
| rm | Remove file(s) or directory(s) |
| rmdir | Remove directory |
| rpm | Query an RPM package file |
| rpm2cpio | Extract the cpio payload from an RPM package |
| rtcwake | Arm the RTC to wake the system |
| script | Record a command's output to a typescript |
| scriptreplay | Replay a typescript using its timing file |
| sddf | Search & Delete Duplicated File |
| sed | Stream editor for filtering and transforming text |
| seq | Print a column of numbers |
| serial | Rename the file to the name with a serial number |
| setarch | Run a program with a changed architecture personality |
| setpriv | Run a program with different privilege settings |
| setsid | Run a program in a new session |
| sh | Command interpreter (MimixBox mbsh compatibility front-end) |
| sha1sum | Calculate or Check secure hash 1 algorithm |
| sha256sum | Calculate or Check secure hash 256 algorithm |
| sha384sum | Calculate or Check secure hash 384 algorithm |
| sha3sum | Calculate or Check SHA-3 message digest |
| sha512sum | Calculate or Check secure hash 512 algorithm |
| shred | Overwrite a file to hide its contents |
| shuf | Generate a random permutation of input lines |
| sl | Cure your bad habit of mistyping |
| sleep | Pause for NUMBER seconds(minutes, hours, days) |
| smemcap | Capture /proc memory data into a tar for smem |
| sort | Sort lines of text files |
| speaker | Read text aloud using an installed TTS engine |
| split | Split a file into pieces |
| stat | Display file or file system status |
| strings | Print printable character sequences in files |
| stty | Print or change terminal line settings |
| sum | Checksum and count the blocks in a file (BSD) |
| swapoff | Disable a swap area |
| swapon | Enable a swap area or list active swaps |
| switch_root | Switch to another root and run init |
| sync | Synchronize cached writes to persistent storage |
| sysctl | Read and write kernel parameters at runtime |
| syslogd | Minimal system logging daemon |
| tac | Print the file contents from the end to the beginning |
| tail | Print the last NUMBER(default=10) lines |
| tar | Archive files (create, list, extract) |
| taskset | Set or get a process's CPU affinity |
| tee | Read from standard input and write to standard output and files |
| test | Evaluate a conditional expression |
| time | Run a command and report how long it took |
| timeout | Run a command with a time limit |
| top | Display system summary and top processes |
| touch | Update the access and modification times of each FILE to the current time |
| tr | Translate or delete characters |
| tree | List directory contents in a tree-like format |
| true | Do nothing. Return success(0) |
| truncate | Shrink or extend the size of a file to a given size |
| tsort | Topological sort of a directed graph |
| tty | Print the file name of the terminal connected to stdin |
| tune2fs | Show ext2/ext3/ext4 filesystem parameters |
| umount | Unmount a filesystem |
| uname | Print system information |
| uncompress | Decompress LZW (.Z) files |
| unexpand | Convert N space to TAB(default:N=8) |
| uniq | Report or omit repeated lines |
| unit | BusyBox unit-test runner (not shipped by MimixBox) |
| unix2dos | Change LF to CRLF |
| unlink | Remove a single file by calling the unlink function |
| unlzma | Decompress lzma files |
| unshadow | Combine passwd and shadow files for password auditing |
| unshare | Run a program with unshared namespaces |
| unxz | Decompress xz files |
| unzip | Extract files from a ZIP archive |
| uptime | Tell how long the system has been running |
| users | Print the user names of those currently logged in |
| usleep | Pause for N microseconds |
| uudecode | Decode a uuencoded (or base64) file |
| uuencode | Encode a file for transmission over text channels |
| uuidgen | Print UUID (Universally Unique IDentifier) |
| valid-shell | Verify if /etc/shells is valid |
| vi | A minimal vi-style screen text editor |
| vmstat | Report virtual memory statistics |
| w | Show who is logged on and a system summary |
| wall | Write a message to all logged-in users |
| watch | Execute a program periodically, showing output fullscreen |
| wc | Print newline, word, and byte counts for each file |
| wget | The non-interactive network downloader |
| which | Returns the file path which would be executed in the current environment |
| who | Show who is logged on |
| whoami | Print login user name |
| whris | Show management information for a domain's IP addresses |
| xargs | Build and execute command lines from standard input |
| xxd | Make a hex dump or do the reverse |
| xz | Compress or decompress files (xz) |
| xzcat | Decompress xz data to standard output |
| yes | Output a string repeatedly until killed |
| zcat | Decompress gz data to standard output |
| zip | Package and compress files into a ZIP archive |
| zip-pwcrack | Recover the password of a ZipCrypto-encrypted archive |
<!-- COMMAND_LIST_END -->

## Install

The [Release Page](https://github.com/nao1215/mimixbox/releases) distributes a `tar.gz` archive for each OS/CPU architecture, plus `.deb`, `.rpm`, and `.apk` packages. The archive is named `mimixbox_<version>_<os>_<arch>.tar.gz` and extracts into a directory containing the `mimixbox` binary and a self-contained `installer.sh`. For example, on Linux (amd64):

```shell
$ tar xf mimixbox_0.36.0_linux_amd64.tar.gz
$ cd mimixbox_0.36.0_linux_amd64
$ sudo ./installer.sh
```

The installer copies the binary to `/usr/local/bin` and creates a symlink there for each applet. It resolves everything relative to itself, so it needs no Git checkout. If you prefer to do it by hand, just install the binary and let it create the symlinks:

```shell
$ sudo install -m 0755 mimixbox /usr/local/bin/
$ sudo mimixbox --install /usr/local/bin
```

### Use "go install"

```shell
$ go install github.com/nao1215/mimixbox/cmd/mimixbox@latest
$ sudo mimixbox --install /usr/local/bin
```

## Original commands

MimixBox has its own commands that do not exist in packages like Coreutils.

### fakemovie

Add a movie-style play button to an image (based on [mattn's fakemovie](https://github.com/mattn/fakemovie)). Use `-o` to set the output name, `-p` for a different button style and `-r` to set the radius.

```shell
$ fakemovie lena.png            # writes lena_fake.png
$ fakemovie -p lena.png -o out.png
```

### ghrdc

GitHub Release Downloads Counter: print how many times a repository's release assets were downloaded, via the GitHub API. `-a` shows the count per release and `-t` the total across all releases. It uses the unauthenticated API, so it is limited to 60 calls per hour and public repositories only.

```shell
$ ghrdc nao1215/mimixbox        # latest release
$ ghrdc -t nao1215/mimixbox     # all releases (total)
```

### path

Manipulate a filename path: print its absolute form, canonical form, directory, basename or extension.

```shell
$ path -a path                  # absolute path
$ path -b /etc/systemd/pstore.conf   # basename -> pstore.conf
$ path -d /etc/ssh/ssh_config        # dirname  -> /etc/ssh
$ path -e archive.tar.gz             # extension -> .gz
```

### sddf

Search & Delete Duplicated File: find files with identical content (compared by md5 checksum) and remove the duplicates, keeping the most recent copy. The list of removed files is written to `duplicated-file.sddf` (change the name with `-o`).

```shell
$ sddf .
```

### serial

Rename the files in a directory to a common base name with a serial number, useful for normalizing a directory of images or downloads. Choose the base name with `-n`, put the number as a prefix (`-p`, default) or suffix (`-s`), and preview with `-d`.

```shell
$ serial -n photo .             # photo0, photo1, ... in the current directory
$ serial -d -n photo .          # dry run: print the renames without applying them
```

## Development

### Tools & Libraries

The table below shows the tools used when developing commands in the MimixBox project.

| Tool | Description |
|:-----|:------|
| go-licenses | License management of dependent libraries |
| make | Build, test, release, etc. |
| curl | Install ShellSpec |
| install | Install the MimixBox binary on the system |
| docker | Test MimixBox inside Docker |
| debootstrap | Test MimixBox inside a jail environment |
| shellspec | End-to-end test |
| golangci-lint | Lint Go code |
| libpam0g-dev (pam-devel) | PAM (Pluggable Authentication Modules) library |

On a Debian-based distribution (e.g. Debian／Ubuntu／Kali Linux／Raspberry Pi OS), install the tools with:

```shell
$ sudo apt install build-essential curl git docker.io debootstrap libpam0g-dev
$ go install github.com/google/go-licenses@latest
$ curl -fsSL https://git.io/shellspec | sh -s -- --yes
```

### How to build

```shell
$ git clone https://github.com/nao1215/mimixbox.git
$ cd mimixbox
$ make build
```

### Debugging

Create a Docker environment:

```shell
$ make docker
# Once the image build finishes, you are inside the container.
```

Create a jail environment:

```shell
$ sudo apt install debootstrap            # If debootstrap is not installed on Ubuntu.
$ make build                              # Build the mimixbox binary.
$ sudo make jail                          # Create the jail at /tmp/mimixbox/jail.

$ sudo chroot /tmp/mimixbox/jail /bin/bash    # Dive into the jail.
# mimixbox --full-install /usr/local/bin      # Install MimixBox commands in the jail.
```

## Contributing

Thanks for taking the time to contribute. See [CONTRIBUTING.md](./CONTRIBUTING.md) for details. Contributions are not limited to development; a GitHub Star is also a motivation to keep developing.

[![Star History Chart](https://api.star-history.com/svg?repos=nao1215/mimixbox&type=Date)](https://star-history.com/#nao1215/mimixbox&Date)

## Contact

To report a bug or request a feature, please use [GitHub Issue](https://github.com/nao1215/mimixbox/issues). Sponsorship is also welcome.

## License

The MimixBox project is licensed under the [Apache License 2.0](./LICENSE). It also incorporates portions of third-party code distributed under the MIT License (for example the `nc`, `whris`, and `fakemovie` applets), whose original copyright and license notices are preserved in the corresponding source files.

## Contributors

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tbody>
    <tr>
      <td align="center" valign="top" width="14.28%"><a href="https://debimate.jp/"><img src="https://avatars.githubusercontent.com/u/22737008?v=4?s=75" width="75px;" alt="CHIKAMATSU Naohiro"/><br /><sub><b>CHIKAMATSU Naohiro</b></sub></a><br /><a href="https://github.com/nao1215/mimixbox/commits?author=nao1215" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/polynomialspace"><img src="https://avatars.githubusercontent.com/u/45617594?v=4?s=75" width="75px;" alt="polynomialspace"/><br /><sub><b>polynomialspace</b></sub></a><br /><a href="https://github.com/nao1215/mimixbox/commits?author=polynomialspace" title="Code">💻</a></td>
      <td align="center" valign="top" width="14.28%"><a href="https://github.com/k-avy"><img src="https://avatars.githubusercontent.com/u/81437739?v=4?s=75" width="75px;" alt="Kavya Shukla"/><br /><sub><b>Kavya Shukla</b></sub></a><br /><a href="https://github.com/nao1215/mimixbox/commits?author=k-avy" title="Code">💻</a></td>
    </tr>
  </tbody>
  <tfoot>
    <tr>
      <td align="center" size="13px" colspan="7">
        <img src="https://raw.githubusercontent.com/all-contributors/all-contributors-cli/1b8533af435da9854653492b1327a23a4dbd0a10/assets/logo-small.svg">
          <a href="https://all-contributors.js.org/docs/en/bot/usage">Add your contributions</a>
        </img>
      </td>
    </tr>
  </tfoot>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!
