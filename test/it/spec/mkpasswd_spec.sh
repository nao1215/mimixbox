Describe 'mkpasswd'
    Include loginutils/mkpasswd_test.sh

    It 'hashes with sha-512 and a fixed salt'
        When call TestMkpasswdSha512
        The output should equal '$6$abcdefgh$ltjgWl6579NluT/Vi1nwEvcil.G5Nbc4NiXZaNGStk8PSwGfQv72N2CKPPrVACtLtip/cZ/1GM/O6IND4WQhG.'
        The status should be success
    End
    It 'reads the password from stdin'
        When call TestMkpasswdStdin
        The output should equal '$1$abcdefgh$'
        The status should be success
    End
End
