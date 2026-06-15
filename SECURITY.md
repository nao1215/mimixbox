# Security Policy

## Reporting a Vulnerability

Report security-related issues or vulnerabilities **privately**. Use GitHub's
[private vulnerability reporting](https://github.com/nao1215/mimixbox/security/advisories/new)
for this repository, or email [n.chika156@gmail.com](mailto:n.chika156@gmail.com).
Please include the affected applet(s), the MimixBox version (`mimixbox --version`),
your OS/architecture, and a minimal reproduction.

Do **not** open a public GitHub issue, pull request, or discussion for a
suspected vulnerability until a fix has been released and you have been told it
is safe to disclose.

### Response policy

This is a volunteer-maintained project, so timelines are best-effort:

- Acknowledgement of your report within 7 days.
- An initial assessment (severity and whether it is in scope) within 14 days.
- For confirmed issues, a fix in a new release as soon as practical, followed by
  coordinated public disclosure (typically via a GitHub Security Advisory).

Please allow a reasonable embargo period before disclosing publicly.

## Scope

MimixBox is a collection of Unix command applets in a single binary. Several applets read, write, or delete files (`rm`, `mv`, `cp`, `touch`), change ownership and permissions (`chown`, `chgrp`), send signals (`kill`), or operate on the running system (`chroot`, `halt`, `reboot`, `add-shell`). Running these as root affects the host exactly as the equivalent system commands would. Install MimixBox and create its symbolic links only on systems you control, and review what an applet does before running it with elevated privileges.

## Supported Versions

MimixBox follows a rolling-release model. Only the **latest tagged release** on
the [Release Page](https://github.com/nao1215/mimixbox/releases) is supported;
security fixes are shipped as a new release rather than backported to older
versions. There are no long-term-support branches. Because MimixBox is a
system-level multi-call binary whose applets can run with elevated privileges,
always upgrade to the latest release and reinstall its applet symlinks before
relying on it on a system you administer.

| Version            | Supported          |
|:-------------------|:-------------------|
| Latest release     | :white_check_mark: |
| Older releases     | :x:                |

## Acknowledgments

We thank the security researchers and contributors who responsibly report security issues and work with us to make the project safer for everyone.
