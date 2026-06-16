Describe 'stat GNU flags'
    Include fileutils/stat_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints name and size via --printf with no trailing newline'
        When call TestStatPrintf
        The output should equal "${MIMIXBOX_IT_ROOT}/stat_file=5"
        The status should be success
    End

    It 'interprets backslash escapes in --printf'
        When call TestStatPrintfName
        The output should equal "${MIMIXBOX_IT_ROOT}/stat_file 5"
        The status should be success
    End

    It 'appends a trailing newline for --format'
        When call TestStatFormatNewline
        The output should equal '1'
        The status should be success
    End

    It 'prints a single space-separated terse line'
        When call TestStatTerseFieldCount
        The output should equal '15'
        The status should be success
    End

    It 'reports the size as the second terse field'
        When call TestStatTerseSize
        The output should equal '5'
        The status should be success
    End
End
