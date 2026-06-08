Describe 'speaker'
    Include shellutils/speaker_test.sh
    It 'errors when no text is given'
        When call TestSpeakerNoText
        The output should include 'rc:1'
        The status should be success
    End
End
