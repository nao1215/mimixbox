# GNU install flag parity (Issue #755): --owner/-o, --group/-g, --strip/-s,
# --backup[=CONTROL], and --suffix/-S.
#
# Each example works in its own subdirectory under the per-run temp root so the
# cases stay isolated. Assertions only check behavior that is deterministic for
# a non-root CI user.
Describe 'install GNU flags'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/install_gnu"
        rm -rf "${WORK}"
        mkdir -p "${WORK}"
    }
    cleanup() {
        rm -rf "${MIMIXBOX_IT_ROOT}/install_gnu"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    # --backup=simple moves the existing destination aside to dst~ before
    # overwriting it with the new content.
    It 'makes a simple backup before overwriting with --backup=simple'
        TestBackupSimple() {
            cd "${WORK}" || exit 1
            printf 'new\n' > src
            printf 'old\n' > dst
            install --backup=simple src dst
            cat dst dst~
        }
        When call TestBackupSimple
        The output should equal "$(printf 'new\nold')"
        The status should be success
    End

    # --suffix overrides the simple-backup suffix.
    It 'honors --suffix when backing up'
        TestBackupSuffix() {
            cd "${WORK}" || exit 1
            printf 'new\n' > src
            printf 'old\n' > dst
            install --backup=simple -S .bak src dst
            cat dst.bak
        }
        When call TestBackupSuffix
        The output should equal "old"
        The status should be success
    End

    # numbered backups produce dst.~1~, dst.~2~, ... on successive installs.
    It 'makes numbered backups with --backup=numbered'
        TestBackupNumbered() {
            cd "${WORK}" || exit 1
            printf 'v1\n' > src
            printf 'orig\n' > dst
            install --backup=numbered src dst
            printf 'v2\n' > src
            install --backup=numbered src dst
            cat dst dst.~1~ dst.~2~
        }
        When call TestBackupNumbered
        The output should equal "$(printf 'v2\norig\nv1')"
        The status should be success
    End

    # existing-mode falls back to a simple backup when no numbered backup exists.
    It 'uses a simple backup with --backup=existing when none are numbered'
        TestBackupExisting() {
            cd "${WORK}" || exit 1
            printf 'new\n' > src
            printf 'old\n' > dst
            install --backup=existing src dst
            cat dst~
        }
        When call TestBackupExisting
        The output should equal "old"
        The status should be success
    End

    # -o/-g to uid/gid 0 must fail for a non-root user (EPERM), but the file is
    # still installed first. This spec assumes a non-root CI user.
    It 'attempts chown and fails as non-root with --owner/--group'
        TestOwnerNonRoot() {
            cd "${WORK}" || exit 1
            printf 'data\n' > src
            install -o 0 -g 0 src dst
        }
        When call TestOwnerNonRoot
        The error should include "ownership"
        The status should be failure
        # The destination is written before the chown is attempted.
        The path "${WORK}/dst" should be exist
    End

    # An unknown owner name is rejected.
    It 'rejects an invalid owner name'
        TestInvalidOwner() {
            cd "${WORK}" || exit 1
            printf 'data\n' > src
            install -o no-such-user-xyz src dst
        }
        When call TestInvalidOwner
        The error should include "invalid user"
        The status should be failure
    End

    # An unknown --backup CONTROL word is rejected.
    It 'rejects an invalid --backup control'
        TestInvalidBackup() {
            cd "${WORK}" || exit 1
            printf 'data\n' > src
            install --backup=bogus src dst
        }
        When call TestInvalidBackup
        The error should include "invalid argument"
        The status should be failure
    End
End
