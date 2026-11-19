#!/bin/sh
# gen.sh KIND N OUT writes the input of one benchmark in bench/himorime.yaml.
# Both revisions of a comparison run the working tree's copy
# (${head_root}/gen.sh), so they read the same bytes. Only sh and awk are
# needed, and the contents depend on N alone.
#
#   lines  N lines of words in a fixed pseudo-random order; OUT is a file
#   files  a directory of N files of different sizes; OUT is the directory
set -eu

kind=$1
n=$2
out=$3

case $kind in
lines)
	awk -v n="$n" 'BEGIN {
		x = 12345
		for (i = 0; i < n; i++) {
			x = (x * 1103515245 + 12345) % 2147483648
			printf "%010d word%d %s\n", x, i % 97, (i % 3 ? "alpha" : "beta")
		}
	}' > "$out"
	;;
files)
	rm -rf "$out"
	mkdir -p "$out"
	awk -v n="$n" -v dir="$out" 'BEGIN {
		for (i = 0; i < n; i++) {
			f = dir "/file" i ".txt"
			for (j = 0; j <= i % 7; j++) printf "x" > f
			close(f)
		}
	}'
	;;
*)
	echo "gen.sh: unknown kind $kind" >&2
	exit 2
	;;
esac
