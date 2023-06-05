#!/bin/sh

set -x

goos=$(uname -s | tr "[:upper:]" "[:lower:]")

case $(uname -m) in
  x86_64) goarch="amd64";;
  aarch64) goarch="arm64";;
esac

\curl -SSL "https://github.com/jhoblitt/arborist/releases/latest/download/arborist-${goos}-${goarch}.tar.gz" | tar zxvf -
