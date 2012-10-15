#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export GOPATH=$DIR

cd $DIR/src/github.com/stretchrcom/professor
go build
cd $DIR

./src/github.com/stretchrcom/professor/professor $1 $2