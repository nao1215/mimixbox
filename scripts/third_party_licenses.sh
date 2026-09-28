#!/bin/sh
# [Description]
#  Collect the license texts of every module linked into the shipped mimixbox
#  binary into third_party_licenses/linux/<module path>/. GoReleaser runs this
#  as a before hook and ships the directory as THIRD_PARTY_LICENSES/ in the
#  release archives and packages; `make licenses` runs it locally.
#
#  go-licenses v2 is required (go install github.com/google/go-licenses/v2@v2.0.1).
#  GOROOT is set from `go env GOROOT` because go-licenses recognises the
#  standard library by path prefix, and a toolchain switch (go.mod asking for a
#  newer patch than the installed Go) would otherwise make it treat every
#  standard package as an unlicensed dependency.
#
#  github.com/golang/freetype is dual licensed under the FreeType License or
#  GPL-2.0-or-later, and its LICENSE only points at licenses/ftl.txt and
#  licenses/gpl.txt, so go-licenses cannot classify it. MimixBox uses it under
#  the FreeType License: it is skipped here and its LICENSE and ftl.txt are
#  copied by hand. The README carries the acknowledgement that license asks for.
set -eu

out=third_party_licenses/linux

GOROOT="$(go env GOROOT)"
export GOROOT

GOOS=linux CGO_ENABLED=0 go-licenses save ./cmd/mimixbox \
    --save_path="${out}" --force \
    --ignore github.com/golang/freetype

freetype_dir="$(go list -m -f '{{.Dir}}' github.com/golang/freetype)"
mkdir -p "${out}/github.com/golang/freetype"
cp "${freetype_dir}/LICENSE" "${freetype_dir}/licenses/ftl.txt" "${out}/github.com/golang/freetype/"
