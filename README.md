<div align="center">
<img src="https://github.com/nao1215/mimixbox/blob/main/docs/images/logo.jpg" width="100">
</div>

[![Build](https://github.com/nao1215/mimixbox/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/nao1215/mimixbox/actions/workflows/build.yml)
[![UnitTest](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml/badge.svg?branch=main&event=push)](https://github.com/nao1215/mimixbox/actions/workflows/unit_test.yml)
[![IntegrationTest](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml/badge.svg?event=push)](https://github.com/nao1215/mimixbox/actions/workflows/integration_test.yml)
![GitHub](https://img.shields.io/github/license/nao1215/mimixbox)
![GitHub all releases](https://img.shields.io/github/downloads/nao1215/mimixbox/total)
![Lines of code](https://img.shields.io/tokei/lines/github/nao1215/mimixbox?style=plastic)

[[日本語](docs/introduction/ja/README.MD)]
# MimixBox - mimic BusyBox on Linux
MimixBox has many Unix commands in the single binary like BusyBox. However, MimixBox aim for the different uses from BusyBox. Specifically, it is supposed to be used in the desktop environment, not the embedded environment.  
Also, the MimixBox project maintainer plan to have a wide range of built-in commands (applets) from basic command provided by Coreutils and others to experimental commands.

# Installation.
The source code and binaries are distributed on [the Release Page](https://github.com/nao1215/mimixbox/releases) in ZIP format and tar.gz format. Choose the binary that suits your OS and CPU architecture.
For example, in the case of Linux (amd64), you can install the MimixBox and documents on your system with the following command:

```
$ tar xf mimixbox-0.0.1-linux-arm64.tar.gz
$ cd mimixbox-0.0.1-linux-arm64
$ sudo ./installer.sh
```

If you have a Golang development environment, you can also install it by the following method. This method does not install the documentation.

```
$ go install github.com/nao1215/mimixbox/cmd/mimixbox
$ mimixbox --install /usr/local/bin
```
# Development 
## Tools
The table below shows the tools used when developing the commands in the MimixBox project.
| Tool | description |
|:-----|:------|
| go-licenses | Used for license management of dependent libraries|
| pandoc   | Convert markdown files to manpages |
| make   | Used for build, test, release, etc |
| gzip   | Used for compress man pages |
| curl | Used for install ShellSpec |
| install   | Used for install MimixBox binary and document in the system |
| docker| Used for testing Mimixbox inside Docker|
| debootstrap| Used for testing Mimixbox inside jail envrioment|
|  shellspec   | Used for integration test|  

If you use Debian-based distribution (e.g. Debian／Ubuntu／Kali Linux／Raspberry Pi OS), You can install tools with the following command.

```
$ sudo apt install build-essential curl git pandoc gzip docker.io debootstrap
$ go install github.com/google/go-licenses@latest
$ curl -fsSL https://git.io/shellspec | sh -s -- --yes
```
  
## How to build

```
$ git clone https://github.com/nao1215/mimixbox.git
$ cd mimixbox
$ make build
```

# [Debugging](docs/introduction/en/DebugAndTest.md)
## How to create docker environment
```
$ make docker

※ Once the Docker image build is complete, you'll be inside the container.
$ 
```
## How to create jail environment
``` 
$ sudo apt install debootstrap    ※ If you have not installed debootstrap in Ubuntu.
$ make build                      ※ Build mimixbox binary
$ sudo make jail                  ※ Create jail environment at /tmp/mimixbox/jail

$ sudo chroot /tmp/mimixbox/jail /bin/bash   ※ Dive to jail
# mimixbox --full-install /usr/local/bin     ※ Install MimixBox's command in jail
```

# Roadmap
- Step1. Implements many common Unix commands (〜Version 0.x.x).
- Step2. Increase the options for each commands (〜Version 1.x.x).
- Step3. Change the command to modern specifications(〜Version 2.x.x)
  
Now, MimixBox has not implemented enough commands ([currently supported command list is here](./docs/introduction/en/CommandAppletList.md)). Therefore, as a project, we will increase the number of commands and aim for a state where dog fooding can be done with the highest priority.
    
Then increase the command options to the same level as Coreutils and other packages. Note that MimixBox does not aim to create commands equivalent to Coreutils. However, we think it's better to have the options that Linux users expect.
  
Finally, it's the phase that makes the Mimix Box unique. Extend commands to high functionality, like [bat](https://github.com/sharkdp/bat) and [lsd](https://github.com/Peltoche/lsd), which are reimplementations of cat and ls.

# Original commands in MimixBox
MimixBox has its own commands that don't exist in packages like Coreutils.
|Command (Applet) Name | Description|
|:--|:--|
|[fakemovie](./docs/examples/fakemovie/en/fakemovie.md) | Adds a video playback button to the image|
|[ghrdc](./docs/examples/ghrdc/en/ghrdc.md) | GitHub Relase Download Counter|
|[path](./docs/examples/path/en/path.md) | Manipulate filename path|
|[serial](./docs/examples/serial/en/serial.md) | Rename the file to the name with a serial number|

# Contact
If you would like to send comments such as "find a bug" or "request for additional features" to the developer, please use one of the following contacts.

- [GitHub Issue](https://github.com/nao1215/mimixbox/issues)
- [Twitter@mimixbox156](https://twitter.com/mimixbox156) ※ MimixBox project information

# LICENSE
The MimixBox project is licensed under the terms of the MIT license and Apache License 2.0.  
See [LICENSE](./LICENSE) and [NOTICE](./NOTICE)
