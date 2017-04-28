#!/bin/bash

cd goodies
go build 
go test
cd ..
cd goodies-server
go build
go test
cd ..
cd goodies-cli
go build 
go test
cd ..

exit 0
