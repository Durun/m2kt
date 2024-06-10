#!/bin/bash -e

cd $(git rev-parse --show-toplevel)

if [ -z $1 ]; then
    echo "Usage: GOOS=<os> GOARCH=<arch> $0 [-o <name>] pkg1 pkg2 ..."
    echo "  os:     linux darwin windows"
    echo "  arch:   amd64 arm64"
    exit 1
fi

case $1 in
    -o)
        shift
        OUTNAME=$1
        shift
        ;;
esac

PKGS=$@
GOARCH=${GOARCH:-$(go env GOARCH)}
GOOS=${GOOS:-$(go env GOOS)}

export GOARCH=${GOARCH}
export GOOS=${GOOS}
export CGO_ENABLED=0
LDFLAGS=(
  "-extldflags=-static"

  "-s" # Omit the symbol table and debug information
  "-w" # Omit the DWARF symbol table
)

for pkg in ${PKGS}; do
    DIR="${GOOS}_${GOARCH}"
    OUT_PATH="dist/${DIR}/${OUTNAME:-$(basename ${pkg})}"
    go build -o ${OUT_PATH} "-ldflags=${LDFLAGS[*]}" -trimpath "${pkg}"
    chmod +x ${OUT_PATH}

    echo "${pkg} -> ${OUT_PATH}"
done
