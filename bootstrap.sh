#!/bin/bash

# check if go is installed
go version

if [ $? -ne 0 ]; then
    echo 'Go is not installed or in your $PATH, please install or configure your $PATH'
    exit 1
fi

deps=( "gopkg.in/redis.v2" "gopkg.in/mgo.v2" "gopkg.in/mgo.v2/bson" "github.com/montanaflynn/stats" "golang.org/x/net/html")

echo "Getting project dependencies"
for dep in $deps; do
    echo "Installing $dep..."
    go get $dep
done

cd ./src/youtrend
go build -o ../../youtrend
cd ../..

exit 0
