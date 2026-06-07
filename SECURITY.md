# Security Policy

## Reporting a Vulnerability

If you discover any security-related issues or vulnerabilities, please contact us at [n.chika156@gmail.com](mailto:n.chika156@gmail.com). We appreciate your responsible disclosure and will work with you to address the issue promptly.

Please avoid filing a public issue for security problems until a fix is available.

## Scope

MimixBox is a collection of Unix command applets in a single binary. Several applets read, write, or delete files (`rm`, `mv`, `cp`, `touch`), change ownership and permissions (`chown`, `chgrp`), send signals (`kill`), or operate on the running system (`chroot`, `halt`, `reboot`, `add-shell`). Running these as root affects the host exactly as the equivalent system commands would. Install MimixBox and create its symbolic links only on systems you control, and review what an applet does before running it with elevated privileges.

## Supported Versions

We recommend using the latest release for the most up-to-date and secure experience. Security updates are provided for the latest stable version.

## Acknowledgments

We thank the security researchers and contributors who responsibly report security issues and work with us to make the project safer for everyone.
