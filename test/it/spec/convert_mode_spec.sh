Describe 'dos2unix/unix2dos preserve file mode'
    Include textutils/convert_mode_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'
    It 'dos2unix keeps the original mode'
        When call TestDos2unixKeepsMode
        The output should equal '600'
        The status should be success
    End
    It 'unix2dos keeps the original mode'
        When call TestUnix2dosKeepsMode
        The output should equal '600'
        The status should be success
    End
End
