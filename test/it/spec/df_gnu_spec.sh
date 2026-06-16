Describe 'df GNU output/total/type/block-size/all flags (issue #754)'

    Describe '--output selects and orders columns'
        It 'prints the selected column headers in order'
            When run df --output=source,fstype,size,used,avail,pcent,target
            The status should be success
            The line 1 should include 'Filesystem'
            The line 1 should include 'Type'
            The line 1 should include 'Size'
            The line 1 should include 'Used'
            The line 1 should include 'Avail'
            The line 1 should include 'Use%'
            The line 1 should include 'Mounted on'
        End

        It 'honors a reordered field list'
            When run df --output=target,source
            The status should be success
            # "Mounted on" (target) comes before "Filesystem" (source).
            The line 1 should match pattern 'Mounted on*Filesystem*'
        End

        It 'rejects an unknown field'
            When run df --output=bogus
            The status should be failure
            The stderr should include 'bogus'
        End
    End

    Describe '--total appends a grand-total row'
        It 'emits a row labeled total'
            When run df --total --output=source,size,used,avail,target
            The status should be success
            The output should include 'total'
        End

        It 'works with the classic layout too'
            When run df --total
            The status should be success
            The output should include 'total'
        End
    End

    Describe '--type limits the listing'
        It 'accepts a type filter and exits cleanly'
            # tmpfs is present on essentially every Linux box.
            When run df --type=tmpfs --output=fstype,target
            The status should be success
            The line 1 should include 'Type'
        End

        It 'is repeatable'
            When run df -t tmpfs -t ext4 --output=fstype
            The status should be success
            The line 1 should include 'Type'
        End
    End

    Describe '--block-size scales sizes'
        It 'labels the block-size in the classic header'
            When run df --block-size=1M
            The status should be success
            The line 1 should include '1048576-blocks'
        End

        It 'rejects an invalid size'
            When run df --block-size=1Z
            The status should be failure
            The stderr should include 'block-size'
        End
    End

    Describe '--all includes more filesystems'
        It 'lists at least as many rows with -a as without'
            all_rows=$(df -a --output=target | wc -l)
            base_rows=$(df --output=target | wc -l)
            When call test "$all_rows" -ge "$base_rows"
            The status should be success
        End
    End
End
