Describe 'Word Count without options'
    Include it/textutils/wc_test.sh

    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'says "  6  16 126 /tmp/mimixbox/it/test.txt"'
        When call TestWcWithNoOption
        The output should equal '  6  16 126 /tmp/mimixbox/it/test.txt'
    End
End