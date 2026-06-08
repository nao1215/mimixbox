Describe 'cpio'
    Include archival/cpio_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'round-trips a file through -o and -i'
        When call TestCpioRoundTrip
        The output should equal 'payload'
        The status should be success
    End

    It 'lists archive contents with -i -t'
        When call TestCpioList
        The output should equal 'file.txt'
        The status should be success
    End
End
