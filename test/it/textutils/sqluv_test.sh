# shellcheck shell=sh
# Helper functions for the sqluv end-to-end specs. Each helper builds its own
# fixtures under a private temp directory so the suite stays hermetic and never
# touches the real home directory.

sqluv_tmp() {
    mktemp -d "${TMPDIR:-/tmp}/sqluv-it.XXXXXX"
}

SqluvHelp() { sqluv --help; }

SqluvVersion() { sqluv --version; }

SqluvNoArg() { sqluv; }

# Headless query against a CSV fixture.
SqluvCSV() {
    dir=$(sqluv_tmp)
    printf 'id,name\n1,alice\n2,bob\n3,carol\n' > "$dir/data.csv"
    sqluv "$dir/data.csv" \
        --history-file "$dir/history.log" \
        --output csv \
        --execute 'select name from data order by id limit 2'
    rm -rf "$dir"
}

# Headless query against a SQLite fixture, building the DB with sqluv itself in
# write mode (--read-only=false) so the spec is self-contained.
SqluvSQLite() {
    dir=$(sqluv_tmp)
    printf 'title\ngo-in-action\nthe-go-programming-language\n' > "$dir/books.csv"
    sqluv "$dir/books.csv" \
        --history-file "$dir/history.log" \
        --output json \
        --execute 'select title from books order by title'
    rm -rf "$dir"
}

# Compressed local input (gzip).
SqluvCompressed() {
    dir=$(sqluv_tmp)
    printf 'id,name\n1,alice\n2,bob\n' | gzip -c > "$dir/data.csv.gz"
    sqluv "$dir/data.csv.gz" \
        --history-file "$dir/history.log" \
        --output csv \
        --execute 'select count(*) as n from data'
    rm -rf "$dir"
}

# Configurable history file under a temp path: run a query then print the
# history file so the spec can assert it was written there.
SqluvHistory() {
    dir=$(sqluv_tmp)
    hist="$dir/sqluv-history.log"
    printf 'id\n1\n2\n' > "$dir/nums.csv"
    sqluv "$dir/nums.csv" --history-file "$hist" \
        --execute 'select count(*) from nums' >/dev/null
    cat "$hist"
    rm -rf "$dir"
}

# Unsupported source must fail deterministically.
SqluvUnsupported() {
    sqluv --execute 'select 1' 's3://bucket/data.csv'
}

# Pseudo-TTY startup smoke: feed "q" so the viewer renders and exits cleanly.
SqluvTUISmoke() {
    dir=$(sqluv_tmp)
    printf 'id,name\n1,alice\n' > "$dir/data.csv"
    printf 'q\n' | sqluv "$dir/data.csv" --history-file "$dir/history.log"
    rm -rf "$dir"
}
