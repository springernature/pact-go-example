#!/usr/bin/env bash
set -e -x

echo "Start go-pact daemon"
wget https://github.com/pact-foundation/pact-go/releases/download/v0.0.12/pact-go_linux_amd64.tar.gz
tar xvzf pact-go_linux_amd64.tar.gz
./pact-go daemon &

echo "List whats in the current directory"
ls -lat
echo ""

export GOPATH=$PWD

mkdir -p src/github.com/springernature/
cp -R ./pact-go-example src/github.com/springernature/.
cd src/github.com/springernature/pact-go-example/provider

go test
