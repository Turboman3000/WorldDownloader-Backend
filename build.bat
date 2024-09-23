@echo off

set GOARCH=amd64
set GOOS=linux

go build -o build/wdl-backend src/main.go

copy "Dockerfile" "build/Dockerfile"