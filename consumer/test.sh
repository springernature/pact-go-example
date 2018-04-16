#!/usr/bin/env bash

wget https://github.com/pact-foundation/pact-go/releases/download/v0.0.12/pact-go_linux_amd64.tar.gz
tar xvzf pact-go_linux_amd64.tar.gz
./pact-go daemon &&

go test