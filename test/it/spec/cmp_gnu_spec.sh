Describe 'cmp -n / --bytes'
    Include shellutils/cmp_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports equality when the difference is past the byte limit'
        When call TestCmpBytesLimitEqual
        The output should equal 'rc=0'
        The status should be success
    End

    It 'reports the difference within the byte limit'
        When call TestCmpBytesLimitDiffer
        The output should include 'differ: byte 4, line 1'
        The status should be failure
    End
End

Describe 'cmp -i / --ignore-initial'
    Include shellutils/cmp_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'skips the first N bytes of both files'
        When call TestCmpIgnoreInitial
        The output should equal 'rc=0'
        The status should be success
    End

    It 'skips N bytes of file1 and M of file2 with N:M'
        When call TestCmpIgnoreInitialPair
        The output should include 'differ: byte 3, line 1'
        The status should be failure
    End
End

Describe 'cmp -b / --print-bytes'
    Include shellutils/cmp_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'prints the differing byte values in the message'
        When call TestCmpPrintBytes
        The output should include 'differ: byte 7, line 2 is 163 s 123 S'
        The status should be failure
    End
End
