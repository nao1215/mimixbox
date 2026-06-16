Describe 'wc --total=only prints just the combined total'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'
    It 'says "4 4 8"'
        When call TestWcTotalOnly
        The output should equal "4 4 8"
        The status should be success
    End
End

Describe 'wc --total=never suppresses the total line'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'

    result() {
        printf '%s\n' \
          "3 3 6 ${MIMIXBOX_IT_ROOT}/wc_a.txt" \
          "1 1 2 ${MIMIXBOX_IT_ROOT}/wc_b.txt"
    }
    It 'prints only the per-file rows'
        When call TestWcTotalNever
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'wc --total=always prints a total even for one file'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'

    result() {
        printf '%s\n' \
          "3 3 6 ${MIMIXBOX_IT_ROOT}/wc_a.txt" \
          "3 3 6 total"
    }
    It 'prints the file row and a total row'
        When call TestWcTotalAlways
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'wc --files0-from reads a NUL-separated list of names'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'

    result() {
        printf '%s\n' \
          "3 3 6 ${MIMIXBOX_IT_ROOT}/wc_a.txt" \
          "1 1 2 ${MIMIXBOX_IT_ROOT}/wc_b.txt" \
          "4 4 8 total"
    }
    It 'counts every file named in the list file'
        When call TestWcFiles0From
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'wc --files0-from=- reads the name list from standard input'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'

    result() {
        printf '%s\n' \
          "3 3 6 ${MIMIXBOX_IT_ROOT}/wc_a.txt" \
          "1 1 2 ${MIMIXBOX_IT_ROOT}/wc_b.txt" \
          "4 4 8 total"
    }
    It 'counts the files piped as a NUL-separated list'
        When call TestWcFiles0FromStdin
        The output should equal "$(result)"
        The status should be success
    End
End

Describe 'wc --files0-from combines with --total=only'
    Include textutils/wc_test.sh
    BeforeEach 'SetupWcGnu'
    AfterEach 'CleanupWcGnu'
    It 'says "4 4 8"'
        When call TestWcFiles0FromOnly
        The output should equal "4 4 8"
        The status should be success
    End
End
