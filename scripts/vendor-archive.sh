#!/usr/bin/env bash
# Write .release/bluefin-cli-<version>-vendor.tar.gz: the vendor/ tree of every
# module go.sum pins, for packagers who build from source without network
# access (distribution RPMs, such as projectbluefin/utah-packages). Unpacked
# next to the release's source tree, `go build` uses it as is.
#
# GoReleaser runs this as a before hook and attaches the archive to the release
# and to checksums.txt. The archive is reproducible: sorted names, fixed owner
# and mtime, and gzip without a name or timestamp, so one commit always yields
# the same bytes.
set -euo pipefail

version="${1:?usage: vendor-archive.sh <version> [out-dir]}"
out_dir="${2:-.release}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

go mod vendor -o "${work}/vendor"
mkdir -p "${out_dir}"
tar --sort=name --mtime=@0 --owner=0 --group=0 --numeric-owner --format=gnu \
    -C "${work}" -cf - vendor |
    gzip -9n >"${out_dir}/bluefin-cli-${version}-vendor.tar.gz"
