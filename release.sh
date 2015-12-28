#! /bin/bash

VERSION="${1}"

if [ -z "${VERSION}" ]; then
    echo "specify the version (no `v`)"
    exit 1
fi

set -e

git tag v$VERSION

REV=`git rev-parse v$VERSION`

importpath=github.com/djmitche/proj/internal
go build -o proj \
    -ldflags "-X ${importpath}.Version=${VERSION} -X ${importpath}.Revision=${REV}" \
    main.go
./proj -V

echo "If everything looks good:"
echo " * git push --tags"
echo " * upload proj as the Proj-$VERSION release on github"
