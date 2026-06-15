Describe 'makedevs'
    Include embedded/makedevs_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'creates the directory and file tree from a device table'
        When call TestMakedevsTree
        The output should equal '1'
        The status should be success
    End

    It 'fails without the -d table option'
        When call TestMakedevsUsage
        The status should be failure
        The error should include 'usage: makedevs'
    End

    It 'prints usage for --help'
        When call TestMakedevsHelp
        The output should include 'Usage: makedevs'
        The status should be success
    End

    It 'prints the version line for --version'
        When call TestMakedevsVersion
        The output should include 'makedevs (mimixbox)'
        The status should be success
    End
End
