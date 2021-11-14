#!/bin/bash
ROOT_DIR=$(git rev-parse --show-toplevel)
MAN_DIR="${ROOT_DIR}/docs/man"
MAN_MD=$(find ${MAN_DIR} -name "*.md")

for i in ${MAN_MD}
do
    dir=$(dirname $i)
    base_with_ext=$(basename $i)
    base="${base_with_ext%.*}"

    # make man p(e.g. csv.1.md -> csv.1 -> cav.1.gz)    
    echo "Make man-page: $dir/$base"
    pandoc $i -s -t man > $dir/$base
    echo "Gzip man-page: $dir/${base}.gz"
    gzip -f $dir/$base
done