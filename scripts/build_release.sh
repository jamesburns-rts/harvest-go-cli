#!/bin/bash -e

cd `dirname $0`/..

OUTPUT=harvest.exe
export GOOS=windows 
export GOARCH=386 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT

export GOARCH=amd64 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT

OUTPUT=harvest
export GOOS=darwin 
export GOARCH=386 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT

export GOARCH=amd64 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT

export GOOS=linux 
export GOARCH=386 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT

export GOARCH=amd64 
go build -o $OUTPUT
zip bin/${GOOS}_$GOARCH.zip $OUTPUT
rm $OUTPUT
