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

Describe 'Make directory without arguments'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
      #|Usage:
      #|  mkdir [OPTIONS] PATH
      #|
      #|Application Options:
      #|  -p, --parents  No error if existing, make parent directories as needed
      #|  -v, --version  Show mkdir command version
      #|
      #|Help Options:
      #|  -h, --help     Show this help message
    }

    It 'shows help message'
        When call TestMkdirNoArg
        The stdout should equal "$(result)"
        The status should be failure
    End
End

Describe 'Make directory with --parents option and no arguments'
    Include fileutils/mkdir_test.sh
    BeforeEach 'Setup'
    AfterEach 'Cleanup'

    result() { %text
      #|Usage:
      #|  mkdir [OPTIONS] PATH
      #|
      #|Application Options:
      #|  -p, --parents  No error if existing, make parent directories as needed
      #|  -v, --version  Show mkdir command version
      #|
      #|Help Options:
      #|  -h, --help     Show this help message
    }

    It 'shows help message'
        When call TestMkdirNoArgWithParentsOption
        The stdout should equal "$(result)"
        The status should be failure
    End
End
