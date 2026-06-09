Describe 'mbsh'
    Include shellutils/mbsh_test.sh

    It 'runs an external command and shows a cwd-aware prompt'
        When call TestMbshEcho
        The output should include 'hello'
        The output should include 'mbsh:'
        The status should be success
    End
    It 'ignores comment lines'
        When call TestMbshComment
        The output should include 'ok'
        The status should be success
    End
    It 'expands $? to the last exit status'
        When call TestMbshLastStatus
        The output should include 'status=1'
        The status should be success
    End
    It 'lets a stdin-consuming command read the remaining script input'
        When call TestMbshCatConsumesInput
        The output should include 'hello'
        The status should be success
    End
    It 'does not reparse command-consumed stdin as later commands'
        When call TestMbshNoReparse
        The output should equal 'ok'
        The status should be success
    End
    It 'keeps double-quoted spaces in one argument'
        When call TestMbshDoubleQuote
        The output should include 'a b'
        The status should be success
    End
    It 'expands $HOME'
        When call TestMbshVarExpand
        The output should equal 'expanded'
        The status should be success
    End
    It 'passes a NAME=value prefix to the command environment'
        When call TestMbshEnvAssignment
        The output should include 'FOO=bar'
        The status should be success
    End
End

Describe 'mbsh cd'
    Include shellutils/mbsh_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'changes directory with cd'
        When call TestMbshCd
        The output should include '/tmp/mimixbox/it/mbsh'
        The status should be success
    End
End
