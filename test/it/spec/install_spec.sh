Describe 'install copies a file'
    Include shellutils/install_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'copies the file content'
        When call TestInstallCopyContent
        The output should equal 'hello'
        The status should be success
    End

    It 'sets the requested mode'
        When call TestInstallMode
        The output should equal '640'
        The status should be success
    End

    It 'creates directories with -d'
        When call TestInstallDirectory
        The output should equal 'ok'
        The status should be success
    End

    It 'fails without a destination'
        When call TestInstallNoDest
        The status should be failure
        The error should include 'missing destination'
    End
End
