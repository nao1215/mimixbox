Describe 'readprofile'
    Include util-linux/readprofile_test.sh

    It 'describes itself with --help'
        When call TestReadprofileHelp
        The status should be success
        The output should include 'Usage: readprofile'
        The output should include 'profiling'
    End
End
