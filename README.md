# `gobincache`

`gobincache` checks whether a given install Go binary is up to date with the
version in your projects `go.mod`. This can help remove unecessary work of
installing vendored or cached binaries locally or in CI.

`gobincache` returns an exit code of `2` when the binary binary requires
updating, `0` if it doesn't, and `1` for any other error encountered when
running.

## Install

```bash
go install github.com/CGA1123/gobincache@latest
```

## Usage

```bash
#!/bin/bash

set -u

binary="${1}" # e.g. bin/sqlc
module="${2}" # e.g. github.com/kyleconroy/sqlc/cmd/sqlc

gobincache "${binary}"
status=$?

if [[ $status = "0" ]]; then
  echo "binary is up to date!"
  exit 0
fi

if [[ $status = "1" ]]; then
  echo "there was an error running gobincache"
  exit 1
fi

if [[ $status != "2" ]]; then
  echo "unknown exit code from gobincache"
  exit 1
fi

GOBIN="${PWD}/bin" go install "${module}"
```
