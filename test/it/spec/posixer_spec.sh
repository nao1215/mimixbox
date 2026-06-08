Describe 'posixer'
    Include shellutils/posixer_test.sh
    It 'prints a table header'
        When call TestPosixerHeader
        The output should include 'NAME'
        The output should include 'INSTALLED'
        The status should be success
    End
End
