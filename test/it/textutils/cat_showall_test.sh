SetupShowAll() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    export TEST_FILE_CAT_NP=${MIMIXBOX_IT_ROOT}/cat_nonprinting.bin
    export LANG=C

    mkdir -p ${TEST_DIR}

    # Fixture containing a tab, a blank line, and non-printable bytes
    # (0x01, 0x7f, 0x80). printf writes the raw bytes verbatim.
    printf 'a\tb\x01\n\n\x7f\x80\n' > ${TEST_FILE_CAT_NP}
}

CleanUpShowAll() {
    export TEST_DIR=${MIMIXBOX_IT_ROOT}
    rm -rf ${TEST_DIR}
}

# TestCatShowAllAlias compares `cat -A` with `cat --show-all` on the fixture and
# prints "identical" only when the two outputs are byte-for-byte equal. The
# outputs are written to files and compared with cmp so raw bytes are preserved.
TestCatShowAllAlias() {
    export TEST_FILE_CAT_NP=${MIMIXBOX_IT_ROOT}/cat_nonprinting.bin
    short=${MIMIXBOX_IT_ROOT}/showall_short.out
    long=${MIMIXBOX_IT_ROOT}/showall_long.out
    cat -A ${TEST_FILE_CAT_NP} > ${short}
    cat --show-all ${TEST_FILE_CAT_NP} > ${long}
    if cmp -s ${short} ${long}; then
        echo "identical"
    else
        echo "different"
    fi
}

# TestCatShowNonprintingAlias compares `cat -v` with `cat --show-nonprinting`.
TestCatShowNonprintingAlias() {
    export TEST_FILE_CAT_NP=${MIMIXBOX_IT_ROOT}/cat_nonprinting.bin
    short=${MIMIXBOX_IT_ROOT}/shownp_short.out
    long=${MIMIXBOX_IT_ROOT}/shownp_long.out
    cat -v ${TEST_FILE_CAT_NP} > ${short}
    cat --show-nonprinting ${TEST_FILE_CAT_NP} > ${long}
    if cmp -s ${short} ${long}; then
        echo "identical"
    else
        echo "different"
    fi
}

# TestCatShowAllRendered checks the actual -A rendering: tab -> ^I, non-printing
# bytes via -v notation, and $ at each line end.
TestCatShowAllRendered() {
    export TEST_FILE_CAT_NP=${MIMIXBOX_IT_ROOT}/cat_nonprinting.bin
    cat --show-all ${TEST_FILE_CAT_NP}
}

# TestCatShowNonprintingRendered checks the -v rendering: TAB is left untouched,
# control chars become ^X, DEL becomes ^?, and high bytes get M- notation.
TestCatShowNonprintingRendered() {
    export TEST_FILE_CAT_NP=${MIMIXBOX_IT_ROOT}/cat_nonprinting.bin
    cat --show-nonprinting ${TEST_FILE_CAT_NP}
}
