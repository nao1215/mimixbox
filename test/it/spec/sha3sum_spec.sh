Describe 'sha3sum'
    Include textutils/checksum_test.sh

    It 'defaults to SHA3-256'
        When call TestSha3Default
        The output should equal 'b314e28493eae9dab57ac4f0c6d887bddbbeb810e900d818395ace558e96516d'
        The status should be success
    End
    It 'selects SHA3-512 with -a'
        When call TestSha3_512
        The output should equal 'ac766ba623301e0a'
        The status should be success
    End
End
