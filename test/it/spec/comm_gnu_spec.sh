Describe 'comm --output-delimiter'
    Include textutils/comm_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'separates the columns with the given string'
        When call TestCommOutputDelimiter
        The line 1 of output should equal 'apple'
        The line 2 of output should equal ',,banana'
        The line 3 of output should equal ',,cherry'
        The line 4 of output should equal ',date'
        The status should be success
    End
End

Describe 'comm -z (zero-terminated)'
    Include textutils/comm_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reads and writes NUL-terminated records'
        When call TestCommZeroTerminated
        The output should equal 'banana#'
        The status should be success
    End
End

Describe 'comm --check-order'
    Include textutils/comm_gnu_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports an unsorted input on stderr and fails'
        When call TestCommCheckOrder
        The output should equal 'rc=1'
        The error should include 'file 1 is not in sorted order'
        The status should be success
    End
End
