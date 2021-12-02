Describe 'Make single directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says single'
        When call TestMkdirSingle
        The output should equal 'single'
        The status should be success
    End
End

Describe 'Make parentes/child directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says child'
        When call TestMkdirParent
        The output should equal 'child'
        The status should be success
    End
End

Describe 'Make directory using pipe'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says make directory using pipe'
        When call TestMkdirFromPipe
        The output should equal 'pipe'
        The status should be success
    End
End

Describe 'Make directory without operand'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'print error message'
        When call TestMkdirNoArg
        The error should equal 'mkdir: no operand'
        The status should be failure
    End
End

Describe 'Make directory with --parents option and no operand'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'print error message'
        When call TestMkdirNoArgWithParentsOption
        The error should equal 'mkdir: no operand'
        The status should be failure
    End
End