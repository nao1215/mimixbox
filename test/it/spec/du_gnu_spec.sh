# GNU du flag parity (Issue #753): --max-depth, --exclude, --apparent-size,
# and --one-file-system/-x.
#
# Each example builds its own nested fixture tree under the per-run temp root so
# the cases stay isolated. Sizes are deterministic because every file is written
# with a known byte count.
Describe 'du GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/du_gnu"
        rm -rf "${WORK}"
        mkdir -p "${WORK}/sub/deep"
        # a.txt 1000B -> 1 block, sub/b.txt 2000B -> 2 blocks,
        # sub/deep/c.txt 3000B -> 3 blocks.
        head -c 1000 /dev/zero > "${WORK}/a.txt"
        head -c 2000 /dev/zero > "${WORK}/sub/b.txt"
        head -c 3000 /dev/zero > "${WORK}/sub/deep/c.txt"
    }
    cleanup() {
        rm -rf "${MIMIXBOX_IT_ROOT}/du_gnu"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    # --max-depth=1 prints the operand and its immediate subdirectory, but not
    # the deeper "sub/deep" directory.
    It 'omits directories deeper than --max-depth'
        TestMaxDepth() {
            du --max-depth=1 "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestMaxDepth
        The line 1 of output should equal "$(printf '5\tROOT/sub')"
        The line 2 of output should equal "$(printf '6\tROOT')"
        The status should be success
    End

    # --max-depth=0 collapses the report to the operand total only.
    It 'prints only the operand total with --max-depth=0'
        TestMaxDepthZero() {
            du --max-depth=0 "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestMaxDepthZero
        The output should equal "$(printf '6\tROOT')"
        The status should be success
    End

    # --exclude prunes a matching directory: its subtree is neither listed nor
    # counted toward the operand total (only a.txt remains: 1000B -> 1 block).
    It 'skips entries matching --exclude'
        TestExclude() {
            du --exclude='sub' "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestExclude
        The output should equal "$(printf '1\tROOT')"
        The status should be success
    End

    # --exclude with a glob on the base name skips matching files under -a.
    It 'skips glob-matching files under -a'
        TestExcludeGlob() {
            head -c 4096 /dev/zero > "${WORK}/drop.tmp"
            du -a --exclude='*.tmp' "${WORK}" | grep -c 'drop.tmp' || true
        }
        When call TestExcludeGlob
        The output should equal "0"
        The status should be success
    End

    # --apparent-size reports exact byte totals, which differ from the block
    # count: apparent total is 1000+2000+3000 = 6000 bytes vs 6 blocks default.
    It 'reports exact bytes with --apparent-size'
        TestApparent() {
            du -s --apparent-size "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestApparent
        The output should equal "$(printf '6000\tROOT')"
        The status should be success
    End

    # The default summary still reports 1K blocks (unchanged behaviour).
    It 'reports block counts by default'
        TestDefault() {
            du -s "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestDefault
        The output should equal "$(printf '6\tROOT')"
        The status should be success
    End

    # --one-file-system (-x) does not cross filesystem boundaries; within a
    # single temp filesystem the output matches a plain run.
    It 'matches a plain run on a single filesystem with -x'
        TestOneFS() {
            du -x -s "${WORK}" | sed "s#${WORK}#ROOT#"
        }
        When call TestOneFS
        The output should equal "$(printf '6\tROOT')"
        The status should be success
    End
End
