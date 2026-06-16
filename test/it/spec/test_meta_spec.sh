Describe 'test honors --help/--version only as the sole argument (issue #759)'
    It 'prints help for a sole --help'
        When run env test --help
        The status should be success
        The output should include 'Usage: test'
    End

    It 'prints the version banner for a sole --version'
        When run env test --version
        The status should be success
        The output should include 'test (mimixbox)'
    End

    It 'evaluates an expression when --help is not the sole argument'
        # `test foo = --help` is a string comparison (false), proving --help is an
        # ordinary operand here, not a help request.
        When run env test foo = --help
        The status should be failure
        The output should not include 'Usage: test'
    End
End
