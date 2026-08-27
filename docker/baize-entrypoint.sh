#!/bin/sh
set -eu

umask 077
mkdir -p \
  /opt/reasonix-runtime/go/pkg/mod \
  /opt/reasonix-runtime/go-cache \
  /opt/reasonix-runtime/npm/bin \
  /opt/reasonix-runtime/npm-cache \
  /opt/reasonix-runtime/pnpm \
  /opt/reasonix-runtime/pnpm-store \
  /opt/reasonix-runtime/python-site

exec /usr/local/bin/baize "$@"
