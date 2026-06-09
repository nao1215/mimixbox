# Symlink-dereference contract for cp (GitHub issue #226).
CpSymlinkP() {
    d=$(mktemp -d) || return 1
    echo data > "$d/real.txt"
    ln -s real.txt "$d/link"
    mimixbox cp -P "$d/link" "$d/copy"
    rc=0
    [ -L "$d/copy" ] && echo link || { echo notlink; rc=1; }
    rm -rf "$d"
    return $rc
}

CpSymlinkL() {
    d=$(mktemp -d) || return 1
    echo data > "$d/real.txt"
    ln -s real.txt "$d/link"
    mimixbox cp -L "$d/link" "$d/copy"
    rc=0
    if [ -L "$d/copy" ]; then echo link; rc=1
    elif [ -f "$d/copy" ]; then echo regular
    else echo missing; rc=1
    fi
    rm -rf "$d"
    return $rc
}

CpSymlinkDInTree() {
    d=$(mktemp -d) || return 1
    mkdir -p "$d/src"
    echo data > "$d/src/real.txt"
    ln -s real.txt "$d/src/lnk"
    mimixbox cp -d -r "$d/src" "$d/dst"
    rc=0
    [ -L "$d/dst/lnk" ] && echo link || { echo notlink; rc=1; }
    rm -rf "$d"
    return $rc
}
