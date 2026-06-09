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
    It 'treats an arrow-key escape sequence as a motion, not an edit'
        When call TestViArrowNoAppend
        The output should equal 'hello
world'
        The status should be success
    End
    It 'duplicates a line with yy then p'
        When call TestViYankPaste
        The output should equal 'one
one
two'
        The status should be success
    End
    It 'applies a count to an edit (2x)'
        When call TestViCountedDelete
        The output should equal 'cdef'
        The status should be success
    End
    It 'undoes the last change with u'
        When call TestViUndo
        The output should equal 'keepme'
        The status should be success
    End
    It 'searches with /pattern and moves to the next match with n'
        When call TestViSearchNext
        The output should equal 'x
foo
y
z'
        The status should be success
    End
End
