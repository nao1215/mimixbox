Describe 'self-describing --help'
    Include shellutils/richhelp_test.sh

    It 'cp --help has a description, examples, and exit status'
        When call CpHelp
        The status should be success
        The output should include 'Examples:'
        The output should include 'Exit status:'
    End

    It 'tail --help documents follow mode with examples'
        When call TailHelp
        The status should be success
        The output should include 'Examples:'
        The output should include 'Follow the file'
    End

    It 'wget --help has examples and compatibility notes'
        When call WgetHelp
        The status should be success
        The output should include 'Examples:'
        The output should include 'Notes:'
    End

    It 'mbsh --help describes the shell and its limits'
        When call MbshHelp
        The status should be success
        The output should include 'Examples:'
        The output should include 'Notes:'
    End

    It 'vi --help lists the supported keys'
        When call ViHelp
        The status should be success
        The output should include 'Motions:'
    End

    It 'find --help has examples and notes'
        When call FindHelp
        The status should be success
        The output should include 'Examples:'
        The output should include 'Notes:'
    End
End
