Describe 'Make single directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says single'
        When call TestMkdirSingle
        The output should equal 'single'
    End
End

Describe 'Check status after making single directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says success'
        When call TestMkdirSingleStatus
        The status should be success
    End
End

Describe 'Make three directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1
        #|2
        #|3
    }

    It 'make 1/2/3 directory'
        When call TestMkdirThreeDirectory
        The output should equal "$(result)"
    End
End

Describe 'Check status after making three directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'make 1/2/3 directory'
        When call TestMkdirThreeDirectoryStatus
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
    End
End

Describe 'Check status after making parentes/child directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says success'
        When call TestMkdirParentStatus
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

Describe 'Make directory using pipe'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'
    It 'says success'
        When call TestMkdirFromPipeStatus
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

Describe 'Make three directory. However, can not make one directory'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
        #|1
        #|3
    }

    It 'make 1/3 directory, can not make 2 directory'
        When call TestMkdirThreeDirAndOneIsFail
        The output should equal "$(result)"
        The error should equal "mkdir /mkdir/2: no such file or directory"
    End
End

Describe 'Check status after making three directory( However, can not make one directory)'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    It 'make 1/3 directory, can not make 2 directory'
        When call TestMkdirThreeDirAndOneIsFailStatus
        The error should equal "mkdir /mkdir/2: no such file or directory"
        The status should be failure
    End
End