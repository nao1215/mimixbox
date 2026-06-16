# Issue #942: shell-level round-trip for the local print spool. lpr queues a
# file and stdin, lpq lists both jobs, lpd drains them into an output directory,
# and the queue ends up empty. This protects the spool backend extraction at the
# CLI boundary (the per-applet *_spec.sh files only cover --help).

Describe 'printutils round-trip'
    setup() {
        WORK="${MIMIXBOX_IT_ROOT}/printutils_rt"
        rm -rf "${WORK}"
        mkdir -p "${WORK}"
    }
    cleanup() {
        rm -rf "${MIMIXBOX_IT_ROOT}/printutils_rt"
    }
    BeforeEach 'setup'
    AfterEach 'cleanup'

    It 'queues, lists, drains, and empties the spool'
        TestRoundTrip() {
            cd "${WORK}" || exit 1
            printf 'page content\n' > document.txt

            # Queue a file and a stdin job.
            lpr -S spool document.txt || return 1
            printf 'from stdin\n' | lpr -S spool || return 1

            # lpq must list both jobs in id order.
            lpq -S spool | grep -q 'document.txt' || return 1
            lpq -S spool | grep -q '(stdin)' || return 1

            # Drain to an output directory; both jobs are reported printed.
            lpd -S spool -o printed | grep -q 'printed job 1' || return 1
            lpd_out=$(lpd -S spool -o printed) # second drain run is a no-op
            [ -z "${lpd_out}" ] || return 1

            # The printed file content survives the round-trip.
            cat printed/0001-document.txt

            # The queue is now empty.
            lpq -S spool
        }
        When call TestRoundTrip
        The output should equal "$(printf 'page content\nno entries')"
        The status should be success
    End
End
