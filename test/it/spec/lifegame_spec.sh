Describe 'lifegame CLI contract'
    Include games/lifegame_test.sh
    It 'prints usage with --help and exits 0'
        When call LifegameHelp
        The status should be success
        The output should include 'Usage: lifegame'
    End
End
