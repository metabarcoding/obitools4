#!/bin/bash

# Default values
URL="https://go.dev/dl/"
GITHUB_REPO="https://github.com/metabarcoding/obitools4"
INSTALL_DIR="/usr/local"
OBITOOLS_PREFIX=""
VERSION=""
LIST_VERSIONS=false

# Help message
function display_help {
  echo "Usage: $0 [OPTIONS]"
  echo ""
  echo "Options:"
  echo "  -i, --install-dir       Directory where obitools are installed "
  echo "                          (e.g., use /usr/local not /usr/local/bin)."
  echo "  -p, --obitools-prefix   Prefix added to the obitools command names if you"
  echo "                          want to have several versions of obitools at the"
  echo "                          same time on your system (e.g., -p g will produce "
  echo "                          gobigrep command instead of obigrep)."
  echo "  -v, --version           Install a specific version (e.g., 4.4.8)."
  echo "                          If not specified, installs the latest version."
  echo "  -l, --list              List all available versions and exit."
  echo "  -h, --help              Display this help message."
  echo ""
  echo "Examples:"
  echo "  $0                      # Install latest version"
  echo "  $0 -l                   # List available versions"
  echo "  $0 -v 4.4.8             # Install specific version"
  echo "  $0 -i /opt/local        # Install to custom directory"
}

# List available versions from GitHub releases
function list_versions {
  echo "Fetching available versions..." 1>&2
  echo ""
  curl -s "https://api.github.com/repos/metabarcoding/obitools4/releases" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": "Release_([0-9.]+)".*/\1/' \
    | sort -V -r
}

# Get latest version from GitHub releases
function get_latest_version {
  curl -s "https://api.github.com/repos/metabarcoding/obitools4/releases" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": "Release_([0-9.]+)".*/\1/' \
    | sort -V -r \
    | head -1
}

# Parse command line arguments
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
    -v|--version)
      VERSION="$2"
      shift 2
      ;;
    -l|--list)
      LIST_VERSIONS=true
      shift
      ;;
    -h|--help)
      display_help
      exit 0
      ;;
    *)
      echo "Error: Unsupported option $1" 1>&2
      display_help 1>&2
      exit 1
      ;;
  esac
done

# List versions and exit if requested
if [ "$LIST_VERSIONS" = true ]; then
  echo "Available OBITools4 versions:"
  echo "=============================="
  list_versions
  exit 0
fi

# Determine version to install
if [ -z "$VERSION" ]; then
  echo "Fetching latest version..." 1>&2
  VERSION=$(get_latest_version)
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version" 1>&2
    exit 1
  fi
  echo "Latest version: $VERSION" 1>&2
else
  echo "Installing version: $VERSION" 1>&2
fi

# Construct source URL for the specified version
OBIURL4="${GITHUB_REPO}/archive/refs/tags/Release_${VERSION}.zip"

# The directory from where the script is run
DIR="$(pwd)"

# Create temporary directory
WORK_DIR=$(mktemp -d "obitools4.XXXXXX")

# Check if tmp dir was created
if [[ ! "$WORK_DIR" || ! -d "$WORK_DIR" ]]; then
  echo "Could not create temp dir" 1>&2
  exit 1
fi

mkdir -p "${WORK_DIR}/cache" \
  || (echo "Cannot create ${WORK_DIR}/cache directory" 1>&2
      exit 1)

# Create installation directory
mkdir -p "${INSTALL_DIR}/bin" 2> /dev/null \
  || (echo "Please enter your password for installing obitools in ${INSTALL_DIR}" 1>&2
      sudo mkdir -p "${INSTALL_DIR}/bin")

if [[ ! -d "${INSTALL_DIR}/bin" ]]; then
  echo "Could not create ${INSTALL_DIR}/bin directory for installing obitools" 1>&2
  exit 1
fi

INSTALL_DIR="$(cd ${INSTALL_DIR} && pwd)"

echo "================================" 1>&2
echo "OBITools4 Installation" 1>&2
echo "================================" 1>&2
echo "VERSION=$VERSION" 1>&2
echo "WORK_DIR=$WORK_DIR" 1>&2
echo "INSTALL_DIR=$INSTALL_DIR" 1>&2
echo "OBITOOLS_PREFIX=$OBITOOLS_PREFIX" 1>&2
echo "================================" 1>&2

pushd "$WORK_DIR" > /dev/null || exit

# Detect OS and architecture
OS=$(uname -a | awk '{print $1}')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]] ; then
    ARCH="amd64"
fi

if [[ "$ARCH" == "aarch64" ]] ; then
    ARCH="arm64"
fi

# Download and install Go
echo "Downloading Go..." 1>&2
GOFILE=$(curl -s "$URL" \
            | grep 'class="download"' \
            | grep "\.tar\.gz" \
            | sed -E 's@^.*/dl/(go[1-9].+\.tar\.gz)".*$@\1@' \
            | grep -i "$OS" \
            | grep -i "$ARCH" \
            | head -1)

GOURL=$(curl -s "${URL}${GOFILE}" \
        | sed -E 's@^.*href="(.*\.tar\.gz)".*$@\1@')

echo "Installing Go from: $GOURL" 1>&2

curl -s "$GOURL" | tar zxf -

PATH="$(pwd)/go/bin:$PATH"
export PATH
GOPATH="$(pwd)/go"
export GOPATH
export GOCACHE="$(pwd)/cache"

echo "GOCACHE=$GOCACHE" 1>&2
mkdir -p "$GOCACHE"

# Download OBITools4 source
echo "Downloading OBITools4 v${VERSION}..." 1>&2
echo "Source URL: $OBIURL4" 1>&2

if ! curl -sL "$OBIURL4" > obitools4.zip; then
  echo "Error: Could not download OBITools4 version ${VERSION}" 1>&2
  echo "Please check that this version exists with: $0 --list" 1>&2
  exit 1
fi

unzip -q obitools4.zip

# Find the extracted directory
OBITOOLS_DIR=$(ls -d obitools4-* 2>/dev/null | head -1)

if [ -z "$OBITOOLS_DIR" ] || [ ! -d "$OBITOOLS_DIR" ]; then
  echo "Error: Could not find extracted OBITools4 directory" 1>&2
  exit 1
fi

echo "Building OBITools4..." 1>&2
cd "$OBITOOLS_DIR" || exit
mkdir -p vendor

# Build with or without prefix
if [[ -z "$OBITOOLS_PREFIX" ]] ; then
  make GOFLAGS="-buildvcs=false"
else
  make GOFLAGS="-buildvcs=false" OBITOOLS_PREFIX="${OBITOOLS_PREFIX}"
fi

# Install binaries
echo "Installing binaries to ${INSTALL_DIR}/bin..." 1>&2
(cp build/* "${INSTALL_DIR}/bin" 2> /dev/null) \
   || (echo "Please enter your password for installing obitools in ${INSTALL_DIR}" 1>&2
       sudo cp build/* "${INSTALL_DIR}/bin")

popd > /dev/null || exit

# Cleanup
echo "Cleaning up..." 1>&2
chmod -R +w "$WORK_DIR"
rm -rf "$WORK_DIR"

echo "" 1>&2
echo "================================" 1>&2
echo "OBITools4 v${VERSION} installed successfully!" 1>&2
echo "Binaries location: ${INSTALL_DIR}/bin" 1>&2
if [[ -n "$OBITOOLS_PREFIX" ]] ; then
  echo "Command prefix: ${OBITOOLS_PREFIX}" 1>&2
fi
echo "================================" 1>&2
