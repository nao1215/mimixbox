Describe 'printf honors first-argument --help/--version (issue #758)'
    It 'prints help for a leading --help'
        When run env printf --help
        The status should be success
        The output should include 'Examples:'
        The line 1 of output should start with 'Usage: printf'
    End

    It 'prints the version banner for a leading --version'
        When run env printf --version
        The status should be success
        The output should include 'printf (mimixbox)'
    End

    It 'treats a later --help as an ordinary operand'
        When run env printf 'foo --help\n'
        The status should be success
        The output should equal 'foo --help'
    End
End
