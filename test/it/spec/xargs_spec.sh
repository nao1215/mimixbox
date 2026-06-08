Describe 'xargs'
    Include findutils/xargs_test.sh

    It 'appends stdin items to the command'
        When call TestXargsEcho
        The output should equal 'a b c'
        The status should be success
    End

    It 'splits into groups with -n'
        When call TestXargsMaxArgs
        The output should equal '2'
        The status should be success
    End

    It 'substitutes with -I'
        When call TestXargsReplace
        The output should equal 'hello world'
        The status should be success
    End
End
