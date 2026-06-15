Describe 'setfattr'
    Include embedded/setfattr_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'sets an attribute that getfattr can read back (or skips without xattr support)'
        When call TestSetfattrSetThenRead
        The output should include 'user.k'
        The status should be success
    End

    It 'rejects mutually exclusive -n and -x'
        When call TestSetfattrBadArgs
        The status should be failure
        The error should include 'mutually exclusive'
    End

    It 'prints usage for --help'
        When call TestSetfattrHelp
        The output should include 'Usage: setfattr'
        The status should be success
    End

    It 'prints the version line for --version'
        When call TestSetfattrVersion
        The output should include 'setfattr (mimixbox)'
        The status should be success
    End
End
