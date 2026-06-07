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

The [command (applet) list](./docs/introduction/en/CommandAppletList.md) shows what is currently available. You can also run `mimixbox --list` to print it on the terminal.

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
