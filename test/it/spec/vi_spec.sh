Describe 'vi'
    Include editors/vi_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'deletes a character and writes the file'
        When call TestViDeleteChar
        The output should equal 'ello
world'
        The status should be success
    End
    It 'inserts text and writes the file'
        When call TestViInsert
        The output should equal 'foobar'
        The status should be success
    End
    It 'creates a new file'
        When call TestViNewFile
        The output should equal 'created'
        The status should be success
    End
End
