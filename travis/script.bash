#!/usr/bin/env bash
set -eux

go test -p=1 -v -race ./...
