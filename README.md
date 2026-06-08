<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
![Coverage](https://raw.githubusercontent.com/nao1215/octocovs-central-repo/main/badges/nao1215/mimixbox/coverage.svg)
[![Build](https://github.com/nao1215/mimixbox/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nao1215/mimixbox/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml/badge.svg?branch=main&event=push)](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml)
[![IntegrationTest](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml/badge.svg?event=push)](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml)
![GitHub](https://img.shields.io/github/license/nao1215/mimixbox)
![GitHub all releases](https://img.shields.io/github/downloads/nao1215/mimixbox/total)

# MimixBox - mimic BusyBox on Linux

MimixBox packs many Unix commands into a single binary, like BusyBox. Unlike BusyBox, it targets the desktop environment rather than embedded systems. The project aims for a wide range of built-in commands (applets), from the basics provided by Coreutils to its own experimental commands.

## Commands

The list below is generated from the registered applets by `make command-list`, so it never drifts from the binary. You can also run `mimixbox --list` to print it on the terminal.

<!-- COMMAND_LIST_START -->
There are 98 commands. Run `mimixbox --list` to see them on the terminal.

| Command | Description |
|:--|:--|
| add-shell | Add shell name to /etc/shells |
| ar | Create, modify and extract from archives |
| base64 | Base64 encode/decode from FILR(or STDIN) to STDOUT |
| basename | Print basename (PATH without "/") from file path |
| bunzip2 | Decompress bzip2 (.bz2) files |
| cal | Display a calendar |
| cat | Concatenate files and print on the standard output |
| chgrp | Change the group of each FILE to GROUP |
| chmod | Change file mode bits |
| chown | Change the owner and/or group of each FILE to OWNER and/or GROUP |
| chroot | Run command or interactive shell with special root directory |
| clear | Clear terminal |
| cmp | Compare two files byte by byte |
| cowsay | Print message with cow's ASCII art |
| cp | Copy file(s) otr Directory(s) |
| cpio | Copy files to and from archives |
| cut | Remove sections from each line of files |
| date | Print or set the system date and time |
| dd | Convert and copy a file |
| df | Report file system disk space usage |
| diff | Compare files line by line |
| dirname | Print only directory path |
| dos2unix | Change CRLF to LF |
| du | Estimate file space usage |
| echo | Display a line of text |
| env | Run a program in a modified environment / print the environment |
| expand | Convert TAB to N space (default:N=8) |
| expr | Evaluate expressions |
| fakemovie | Adds a video playback button to the image |
| false | Do nothing. Return unsuccess(1) |
| find | Search for files in a directory hierarchy |
| ghrdc | GitHub Relase Download Counter |
| grep | Print lines that match patterns |
| groups | Print the groups to which USERNAME belongs |
| gunzip | Decompress gzip (.gz) files |
| gzip | Compress or uncompress FILEs (by default, compress FILES in-place) |
| halt | Halt the system |
| head | Print the first NUMBER(default=10) lines |
| hostid | Print the numeric identifier (in hexadecimal) for the current host |
| id | Print User ID and Group ID |
| install | Copy files and set attributes |
| ischroot | Detect if running in a chroot |
| kill | Kill process or send signal to process |
| lifegame | Life game (Conway's Game of Life) |
| ln | Create hard or symbolic link |
| mbsh | Mimix Box Shell |
| md5sum | Calculate or Check md5sum message digest |
| mkdir | Make directories |
| mkfifo | Make FIFO (named pipe) |
| mknod | Make block or character special files |
| mktemp | Create a temporary file or directory |
| mv | Rename SOURCE to DESTINATION, or move SOURCE(s) to DIRECTORY |
| nl | Write each FILE to standard output with line numbers added |
| od | Dump files in octal and other formats |
| path | Manipulate filename path |
| poweroff | Power off the system |
| printenv | Print environment variable |
| printf | Formats and print data |
| pwd | Print Working Directory |
| realpath | Print the resolved absolute file name |
| reboot | Reboot the system |
| remove-shell | Remove shell name from /etc/shells |
| reset | Reset terminal |
| resize | Print commands to set the terminal size |
| rm | Remove file(s) or directory(s) |
| rmdir | Remove directory |
| sddf | Search & Delete Duplicated File |
| sed | Stream editor for filtering and transforming text |
| seq | Print a column of numbers |
| serial | Rename the file to the name with a serial number |
| sha1sum | Calculate or Check secure hash 1 algorithm |
| sha256sum | Calculate or Check secure hash 256 algorithm |
| sha512sum | Calculate or Check secure hash 512 algorithm |
| sl | Cure your bad habit of mistyping |
| sleep | Pause for NUMBER seconds(minutes, hours, days) |
| sort | Sort lines of text files |
| sync | Synchronize cached writes to persistent storage |
| tac | Print the file contents from the end to the beginning |
| tail | Print the last NUMBER(default=10) lines |
| tar | Archive files (create, list, extract) |
| tee | Read from standard input and write to standard output and files |
| test | Evaluate a conditional expression |
| touch | Update the access and modification times of each FILE to the current time |
| tr | Translate or delete characters |
| true | Do nothing. Return success(0) |
| unexpand | Convert N space to TAB(default:N=8) |
| uniq | Report or omit repeated lines |
| unix2dos | Change LF to CRLF |
| unzip | Extract files from a ZIP archive |
| uuidgen | Print UUID (Universal Unique IDentifier |
| valid-shell | Verify if /etc/shells is valid |
| wc | Print newline, word, and byte counts for each file |
| wget | The non-interactive network downloader |
| which | Returns the file path which would be executed in the current environment |
| who | Show who is logged on |
| whoami | Print login user name |
| xargs | Build and execute command lines from standard input |
| zip | Package and compress files into a ZIP archive |
<!-- COMMAND_LIST_END -->

## Install

The release page distributes source code and binaries in ZIP and tar.gz format. Pick the binary that matches your OS and CPU architecture from [the Release Page](https://github.com/nao1215/mimixbox/releases). For example, on Linux (amd64):

```shell
$ tar xf mimixbox-0.30.0-linux-amd64.tar.gz
$ cd mimixbox-0.30.0-linux-amd64
$ sudo ./installer.sh
```

### Use "go install"

```shell
$ go install github.com/nao1215/mimixbox/cmd/mimixbox@latest
$ sudo mimixbox --install /usr/local/bin
```

### Use homebrew

```shell
$ brew install nao1215/tap/mimixbox
```

## Original commands

MimixBox has its own commands that do not exist in packages like Coreutils.

| Command (Applet) Name | Description |
|:--|:--|
| [fakemovie](./docs/examples/fakemovie/en/fakemovie.md) | Adds a video playback button to the image |
| [ghrdc](./docs/examples/ghrdc/en/ghrdc.md) | GitHub Release Download Counter |
| [path](./docs/examples/path/en/path.md) | Manipulate filename path |
| [sddf](./docs/examples/sddf/en/sddf.md) | Search & Delete Duplicated File |
| [serial](./docs/examples/serial/en/serial.md) | Rename the file to a name with a serial number |

## Roadmap

- Step 1. Implement many common Unix commands (〜Version 0.x.x).
- Step 2. Increase the options for each command (〜Version 1.x.x).
- Step 3. Move commands toward modern specifications (〜Version 2.x.x).

MimixBox does not yet have enough commands, so the first priority is increasing their number until the project can be dogfooded. Next, command options are brought closer to Coreutils and other packages. MimixBox does not aim to copy Coreutils, but it does aim to provide the options Linux users expect. The final phase makes MimixBox unique by extending commands toward higher functionality, like [bat](https://github.com/sharkdp/bat) and [lsd](https://github.com/Peltoche/lsd), which are reimplementations of cat and ls.

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

See [DebugAndTest.md](docs/introduction/en/DebugAndTest.md) for details.

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

The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0. See [LICENSE](./LICENSE).

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
