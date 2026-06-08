Describe 'zip and unzip'
    Include archival/zip_test.sh
    BeforeEach 'Setup'
    AfterEach 'CleanUp'

    It 'lists a zipped file via unzip -l'
        When call TestZipThenUnzipList
        The output should include 'a.txt'
        The status should be success
    End

    It 'round-trips a file through zip and unzip'
        When call TestZipThenExtract
        The output should equal 'zipped'
        The status should be success
    End
End
