Describe 'du -b summarizes apparent bytes'
    Include shellutils/du_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports the total byte size'
        When call TestDuBytes
        The output should equal "$(printf '150\t/tmp/mimixbox/it/du')"
        The status should be success
    End
End

Describe 'du -s summarizes in 1K blocks'
    Include shellutils/du_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'reports the total in blocks'
        When call TestDuBlocks
        The output should equal "$(printf '1\t/tmp/mimixbox/it/du')"
        The status should be success
    End
End
