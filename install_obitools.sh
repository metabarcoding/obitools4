#!/bin/bash

INSTALL_DIR="/usr/local"
OBITOOLS_PREFIX=""
# default values
URL="https://go.dev/dl/"
OBIURL4="https://github.com/metabarcoding/obitools4/archive/refs/heads/master.zip"
INSTALL_DIR="/usr/local"
OBITOOLS_PREFIX=""

# help message
function display_help {
  echo "Usage: $0 [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  -i, --install-dir       Directory where obitools are installed "
  echo "                          (as example use /usr/local not /usr/local/bin)."
  echo "  -p, --obitools-prefix   Prefix added to the obitools command names if you"
  echo "                          want to have several versions of obitools at the"
  echo "                          same time on your system (as example -p g will produce "
  echo "                          gobigrep command instead of obigrep)."
  echo "  -h, --help              Display this help message."
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    -i|--install-dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -p|--obitools-prefix)
      OBITOOLS_PREFIX="$2"
      shift 2
      ;;
    -h|--help)
      display_help  1>&2 
      exit 0
      ;;
    *)
      echo "Error: Unsupported option $1"  1>&2
      exit 1
      ;;
  esac
done

# the directory from where the script is run
DIR="$(pwd)"

# the temp directory used, within $DIR
# omit the -p parameter to create a temporal directory in the default location
# WORK_DIR=$(mktemp -d -p "$DIR"  "obitools4.XXXXXX" 2> /dev/null || \
#            mktemp -d -t "$DIR"  "obitools4.XXXXXX")

WORK_DIR=$(mktemp -d "obitools4.XXXXXX")

# check if tmp dir was created
if [[ ! "$WORK_DIR" || ! -d "$WORK_DIR" ]]; then
  echo "Could not create temp dir"  1>&2
  exit 1
fi

mkdir -p "${WORK_DIR}/cache" \
  || (echo "Cannot create ${WORK_DIR}/cache directory" 1>&2  
      exit 1)
      

mkdir -p "${INSTALL_DIR}/bin" 2> /dev/null \
  || (echo "Please enter your password for installing obitools in ${INSTALL_DIR}"  1>&2
      sudo mkdir -p "${INSTALL_DIR}/bin")

if [[ ! -d "${INSTALL_DIR}/bin" ]]; then
  echo "Could not create ${INSTALL_DIR}/bin directory for installing obitools"  1>&2
  exit 1
fi

INSTALL_DIR="$(cd ${INSTALL_DIR} && pwd)"

echo "WORK_DIR=$WORK_DIR"  1>&2
echo "INSTALL_DIR=$INSTALL_DIR"  1>&2
echo "OBITOOLS_PREFIX=$OBITOOLS_PREFIX"  1>&2

pushd "$WORK_DIR"|| exit

OS=$(uname -a | awk '{print $1}')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]] ; then 
    ARCH="amd64" 
fi

if [[ "$ARCH" == "aarch64" ]] ; then 
    ARCH="arm64" 
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

echo "Install GO from : $GOURL" 1>&2
        
curl "$GOURL" \
    | tar zxf -

PATH="$(pwd)/go/bin:$PATH"
export PATH
GOPATH="$(pwd)/go"
export GOPATH

export GOCACHE="$(cd ${WORK_DIR}/cache && pwd)"
echo "GOCACHE=$GOCACHE" 1>&2@
mkdir -p "$GOCACHE"


curl -L "$OBIURL4" > master.zip
unzip master.zip

echo "Install OBITOOLS from : $OBIURL4"

cd obitools4-master || exit
mkdir vendor

if [[ -z "$OBITOOLS_PREFIX" ]] ; then
  make GOFLAGS="-buildvcs=false" 
else
  make GOFLAGS="-buildvcs=false" OBITOOLS_PREFIX="${OBITOOLS_PREFIX}"
fi

(cp build/* "${INSTALL_DIR}/bin" 2> /dev/null) \
   || (echo "Please enter your password for installing obitools in ${INSTALL_DIR}" 
       sudo cp build/* "${INSTALL_DIR}/bin")

popd || exit

chmod -R +w "$WORK_DIR"
rm -rf "$WORK_DIR"

