# GNU cp flag parity (Issues #727, #728, #729): --target-directory/-t and
# --no-target-directory/-T, --parents, and --update/--backup/--suffix.

# Each example works in its own subdirectory under the per-run temp root so the
# cases stay isolated from one another and from the older cp_spec.sh fixtures.
Describe 'cp GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/cp_gnu"
        rm -rf "${WORK}"
        mkdir -p "${WORK}"
    }
    cleanup() {
        rm -rf "${MIMIXBOX_IT_ROOT}/cp_gnu"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    # #727: "cp --target-directory dst a b" must equal "cp a b dst".
    It 'copies into the target directory (-t equals destination-last form)'
        TestTargetDirEquivalence() {
            cd "${WORK}" || exit 1
            printf 'A\n' > a.txt
            printf 'B\n' > b.txt
            mkdir dst_t dst_plain

            cp --target-directory dst_t a.txt b.txt
            cp a.txt b.txt dst_plain

            # Both forms must yield identical destination directory contents.
            diff <(ls dst_t) <(ls dst_plain) && cat dst_t/a.txt dst_t/b.txt
        }
        When call TestTargetDirEquivalence
        The output should equal "$(printf 'A\nB')"
        The status should be success
    End

    # #727: -T must refuse to overwrite an existing directory.
    It 'rejects a directory destination with --no-target-directory (-T)'
        TestNoTargetDir() {
            cd "${WORK}" || exit 1
            printf 'A\n' > a.txt
            mkdir b
            cp -T a.txt b
        }
        When call TestNoTargetDir
        The error should include "cannot overwrite directory"
        The status should be failure
    End

    # #728: --parents recreates the full source prefix under the destination.
    It 'recreates the source path prefix with --parents'
        TestParents() {
            cd "${WORK}" || exit 1
            mkdir -p src/a dst
            printf 'deep\n' > src/a/b.txt
            cp --parents src/a/b.txt dst/
            cat dst/src/a/b.txt
        }
        When call TestParents
        The output should equal "deep"
        The status should be success
    End

    # #729: --backup moves the existing destination aside before overwriting.
    It 'makes a backup before overwriting with --backup'
        TestBackup() {
            cd "${WORK}" || exit 1
            printf 'new\n' > src.txt
            printf 'old\n' > dst.txt
            cp --backup=simple src.txt dst.txt
            # New content lands in dst.txt; the prior content is preserved in dst.txt~.
            cat dst.txt dst.txt~
        }
        When call TestBackup
        The output should equal "$(printf 'new\nold')"
        The status should be success
    End

    # #729: -u skips the copy when the destination is newer than the source.
    It 'skips the copy when the destination is newer (-u)'
        TestUpdate() {
            cd "${WORK}" || exit 1
            # Write src first, then (after a pause) dst, so dst is genuinely
            # newer on disk; -u must therefore leave dst untouched. A short sleep
            # guarantees a distinct mtime without relying on GNU touch -d/-t,
            # which mimixbox's touch does not implement.
            printf 'srcdata\n' > src.txt
            sleep 1.1
            printf 'dstdata\n' > dst.txt
            cp -u src.txt dst.txt
            cat dst.txt
        }
        When call TestUpdate
        The output should equal "dstdata"
        The status should be success
    End
End
