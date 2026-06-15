# MimixBox integration-test (E2E) suite

This directory holds the [ShellSpec](https://shellspec.info/) end-to-end suite
that exercises MimixBox applets as real processes. It is split into two layers:

| Layer | Location | Owns |
| --- | --- | --- |
| **Spec** | `spec/<command>_spec.sh` | `Describe`/`It` assertions on observable behavior. |
| **Helper** | `<subsystem>/<command>_test.sh` | Fixtures, `Setup`/`CleanUp`, and reusable `Test*` invocations a spec `Include`s. |

A spec `Include`s a helper, runs `Setup` in `BeforeEach`, calls the helper's
`Test*` functions via `When call`, and `CleanUp` in `AfterEach`. Keeping fixture
construction in the helper layer means a spec never re-derives temp roots,
sample files, or invocation boilerplate.

## How the suite runs

```sh
make it          # alias for "make test-e2e"
```

`make test-e2e`:

1. builds `mimixbox` and stages every applet as a symlink under
   `test/it/.mbbin` (an isolated PATH directory, git-ignored), then
2. allocates a per-run temp root via `mktemp -d`, exports it as
   `MIMIXBOX_IT_ROOT`, runs `shellspec`, and removes the root on exit.

Because the staged PATH comes first, **bare command names resolve to MimixBox
applets** (e.g. `cat`, `gzip`, `head`). The suite is hermetic: it does not touch
the host's `/usr/bin`.

## Helper conventions (normalized)

Follow `textutils/cat_test.sh` as the canonical pattern. New helpers MUST:

1. **Live at `<subsystem>/<command>_test.sh`.** The subsystem mirrors the applet
   package tree under `internal/applets/<subsystem>/`.
2. **Define `Setup()` and `CleanUp()`** (even if empty: `Setup() { :; }`).
   `Setup` builds fixtures; `CleanUp` removes them.
3. **Define `Test<Command><Case>()` functions** that invoke the applet by its
   bare name and exercise real behavior (round-trips, file transforms, lookups).
4. **Keep every temp file/dir under `${MIMIXBOX_IT_ROOT}`.** This is the per-run
   root from `spec/spec_helper.sh` (env var `MIMIXBOX_IT_ROOT`, convenience
   function `it_root()`). The established idiom is a per-command subdirectory:

   ```sh
   Setup() {
       export TEST_DIR=${MIMIXBOX_IT_ROOT}/<command>
       export LANG=C
       mkdir -p "${TEST_DIR}"
       # ...build fixtures under ${TEST_DIR}...
   }
   CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/<command>"; }
   ```

   **Never hardcode `/tmp/mimixbox`** or any absolute temp path. A per-command
   subdirectory keeps helpers idempotent and safe under repeated/parallel runs.
5. **Set `export LANG=C`** in `Setup` when output is locale-sensitive, so
   assertions are stable.
6. **Re-export needed vars inside each `Test*` function** when a spec calls it
   without `Setup` (matching the `cat_test.sh` pattern), since each ShellSpec
   example may run in its own subshell.
7. **Use MimixBox-compatible flags.** Inside the hermetic PATH, helper commands
   such as `head`/`tail` are also MimixBox applets, so use portable forms
   (`head -n 1`, not `head -1`).

### Minimal template

```sh
# shellcheck shell=sh
Setup() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/foo
    export LANG=C
    mkdir -p "${TEST_DIR}"
    printf 'fixture\n' > "${TEST_DIR}/in.txt"
}
CleanUp() { rm -rf "${MIMIXBOX_IT_ROOT}/foo"; }

TestFooRoundTrip() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}/foo
    foo "${TEST_DIR}/in.txt"
}
```

## Spec-only commands (intentionally no helper)

Many shipped applets are **only safely testable via `--help` / exit-code**: they
need root, a live network peer, real hardware, a daemon/supervisor, or perform a
destructive system action with no deterministic, hermetic, fixture-driven
behavior. For these a `Setup`/`CleanUp` helper would be hollow boilerplate, so
they are **intentionally spec-only** — assert on usage text / exit status in
`spec/<command>_spec.sh` and add no helper. This is an accepted completion of
issue #471.

Categories that are intentionally spec-only:

- **Privileged daemons / supervisors:** `httpd`, `ftpd`, `tftpd`, `telnetd`,
  `dnsd`, `inetd`, `ntpd`, `udhcpd`, `udhcpc`, `udhcpc6`, `dhcprelay`, `tcpsvd`,
  `udpsvd`, `fakeidentd`, `lpd`, `watchdog`, `start-stop-daemon`, `ifplugd`,
  `sendmail`, `popmaildir`.
- **Raw network / interface configuration (root, no deterministic output):**
  `ifconfig`, `ifup`, `ifdown`, `ifenslave`, `ip`, `ipaddr`, `iplink`,
  `ipneigh`, `iproute`, `iprule`, `iptunnel`, `route`, `arp`, `arping`, `brctl`,
  `vconfig`, `nameif`, `zcip`, `slattach`, `tc`, `tunctl`, `ether-wake`,
  `ping6`, `traceroute`, `traceroute6`, `nslookup`, `whois`, `netstat`,
  `netcat`, `nbd-client`, `ftpget`, `ftpput`, `tftp`, `telnet`, `ssl_client`,
  `ssl_server`, `pscan`, `dumpleases`.
- **Kernel module tooling (root / live kernel):** `insmod`, `rmmod`, `lsmod`,
  `modprobe`, `modinfo`, `depmod`.
- **Filesystem creation / block-device tools (root / real device):** `mkdosfs`,
  `mkfs.ext2`, `mkfs.minix`, `mkfs.reiser`, `mkfs.vfat`, `fsck.minix`,
  `partprobe`, `raidautorun`, `readahead`, `resume`, `swapon`, `swapoff`,
  `volname`.
- **System lifecycle / boot (destructive, privileged):** `reboot`, `poweroff`,
  `run-init`, `linuxrc`, `seedrng`.
- **SELinux state mutators (privileged, action-gated):** `chcon`, `runcon`,
  `restorecon`, `setfiles`, `load_policy`, `setenforce`, `setsebool`,
  `getenforce`, `getsebool`, `selinuxenabled`, `sestatus`, `matchpathcon`.
- **Ownership mutators (require root to change owner/group):** `chgrp`, `chown`,
  `addgroup`, `delgroup`.
- **Console / hardware / device tools (real TTY, VT, I2C, mem):** `adjtimex`,
  `conspy`, `microcom`, `openvt`, `loadfont`, `setfont`, `loadkmap`,
  `dumpkmap`, `i2cdetect`, `i2cget`, `i2cset`, `i2cdump`, `devmem`, `rx`.
- **Interactive shells / front-ends (REPL, no deterministic batch output):**
  `sh`, `ash`, `bash`, `hush`, `cttyhack`, `linux32`, `linux64`.
- **Trivial aliases / wrappers already covered elsewhere:** `egrep`, `fgrep`
  (thin `grep` aliases), `busybox` (multi-call dispatcher), `unit` (prints a
  "run go test instead" notice), `dnsdomainname` (host-dependent, may be empty),
  `scriptreplay` (needs a recorded session), `log-collect`, `dpkg`, `dpkg-deb`,
  `pkill`, `pwdx`, `uptime`, `time`, `usleep`, `setsid`, `fsync`, `fallocate`,
  `sl`, `run-parts`, `more`, `less`, `pipe_progress`, `zcat`, `bzcat`, `xzcat`,
  `lzcat`, `lzma`, `unlzma`, `xz`, `bzip2`, `unzip`, `uncompress`, `rpm2cpio`,
  `sum`, `crc32`, `sha384sum`, `sha3sum`, `uudecode`, `uuencode` — these already
  have behavior exercised by an existing helper (e.g. `archival/xzcomp_test.sh`,
  `archival/bunzip2_test.sh`, `archival/compress_test.sh`, `archival/rpm_test.sh`,
  `textutils/checksum_test.sh`, `procps/uptime_pwdx_test.sh`,
  `procps/pgrep_test.sh`, `util-linux/setsid_fallocate_test.sh`,
  `loginutils/run_parts_test.sh`, `console-tools/pager_test.sh`) or are covered
  by a dedicated spec without needing a new fixture helper.

When in doubt, prefer a real helper: only fall back to spec-only when there is
no deterministic behavior reproducible under `${MIMIXBOX_IT_ROOT}` without root,
network, hardware, or a daemon.
