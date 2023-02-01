#!/bin/bash

URL="https://go.dev/dl/"
OBIURL4="https://git.metabarcoding.org/obitools/obitools4/obitools4/-/archive/master/obitools4-master.tar.gz"
PREFIX="/usr/local"

# the directory of the script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# the temp directory used, within $DIR
# omit the -p parameter to create a temporal directory in the default location
WORK_DIR=$(mktemp -d -t "$DIR" "obitools4.XXXXXX")

# check if tmp dir was created
if [[ ! "$WORK_DIR" || ! -d "$WORK_DIR" ]]; then
  echo "Could not create temp dir"
  exit 1
fi

pushd $WORK_DIR

OS=$(uname -a | awk '{print $1}')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]] ; then 
    ARCH="amd64" 
fi

GOFILE=$(curl "$URL" \
            | grep 'class="download"' \
            | grep "\.tar\.gz" \
            | sed -E 's@^.*/dl/(go[1-9].+\.tar\.gz)".*$@\1@' \
            | grep -i "$OS" \
            | grep -i "$ARCH" \
            | head -1)

GOURL=$(curl "${URL}${GOFILE}" \
        | sed -E 's@^.*href="(.*\.tar\.gz)".*$@\1@')
        
curl "$GOURL" \
    | tar zxf -

export PATH="$(pwd)/go/bin:$PATH"

curl "$OBIURL4" \
    | tar zxf - 

cd obitools-master
make

echo "Please enter your password for installing obitools"

sudo mkdir -p "${PREFIX}/bin"
if [[ ! "${PREFIX}/bin" || ! -d "${PREFIX}/bin" ]]; then
  echo "Could not create ${PREFIX}/bin directory for installing obitools"
  exit 1
fi

sudo cp build/* "${PREFIX}/bin"

popd

rm -rf "$WORK_DIR"

