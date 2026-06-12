Describe 'cryptpw'
    Include loginutils/cryptpw_test.sh

    It 'hashes a stdin password with sha-512'
        When call TestCryptpwStdin
        The output should equal '$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.'
        The status should be success
    End
    It 'supports the md5 method'
        When call TestCryptpwMd5
        The output should equal '$1$abcdefgh$'
        The status should be success
    End
End
