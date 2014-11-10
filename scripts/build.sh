#!/usr/bin/env bash

cwd() {
  # Get the parent directory of where this script is.
  local source="${BASH_SOURCE[0]}"
  while [ -h "$source" ] ; do source="$(readlink "$source")"; done
  local dir="$( cd -P "$( dirname "$source" )/.." && pwd )"

  # Change into that directory
  cd $dir

}

configure_os() {
  # If we're building on Windows, specify an extension
  EXTENSION=""
  if [ "$(go env GOOS)" = "windows" ]; then
      EXTENSION=".exe"
  fi

  GOPATHSINGLE=${GOPATH%%:*}
  if [ "$(go env GOOS)" = "windows" ]; then
      GOPATHSINGLE=${GOPATH%%;*}
  fi

  if [ "$(go env GOOS)" = "freebsd" ]; then
    export CC="clang"
  fi

  # On OSX, we need to use an older target to ensure binaries are
  # compatible with older linkers
  if [ "$(go env GOOS)" = "darwin" ]; then
      export MACOSX_DEPLOYMENT_TARGET=10.6
  fi
}

dependencies() {
  # Install dependencies
  echo "--> Installing dependencies to speed up builds..."
  go get \
    -ldflags "${CGO_LDFLAGS}" \
    ./...
}

build() {
  configure_os
  dependencies

  # Get the git commit
  local git_commit=$(git rev-parse HEAD)
  local git_dirty=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)

  # Build!
  echo "--> Building..."
  go build \
      -ldflags "${CGO_LDFLAGS} -X main.GitCommit ${git_commit}${git_dirty}" \
      -v \
      -o bin/hipops${EXTENSION}
  cp bin/hipops${EXTENSION} ${GOPATHSINGLE}/bin
}

main(){
  set -eo pipefail

  cwd

  case "$1" in
  *)    build $@;;
  esac
}

main "$@"
