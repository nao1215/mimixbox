Describe 'getfattr'
    Include embedded/getfattr_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'dumps a user attribute set by setfattr (or skips without xattr support)'
        When call TestGetfattrRoundTrip
        The output should include 'user.demo'
        The status should be success
    End

    It 'fails when no file operand is given'
        When call TestGetfattrNoFile
        The status should be failure
        The error should include 'file operand'
    End

    It 'prints usage for --help'
        When call TestGetfattrHelp
        The output should include 'Usage: getfattr'
        The status should be success
    End

    It 'prints the version line for --version'
        When call TestGetfattrVersion
        The output should include 'getfattr (mimixbox)'
        The status should be success
    End
End
