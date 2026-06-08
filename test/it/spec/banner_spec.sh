Describe 'banner'
    Include jokeutils/banner_test.sh
    It 'prints five rows of art'
        When call TestBanner
        The output should equal '5'
        The status should be success
    End
End
