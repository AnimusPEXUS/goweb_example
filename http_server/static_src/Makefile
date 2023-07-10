export GONOPROXY=github.com/AnimusPEXUS/*
export GOOS=js
export GOARCH=wasm

VERSIONGO=version.go

all: get build

get:
		go get -u -v "./..."
		go mod tidy

build:
		-rm ./$(VERSIONGO)
		echo "package main" >> ./$(VERSIONGO)
		echo "" >> ./$(VERSIONGO)
		echo "const GOWEB_BUILD_TIME=\"$(shell date --utc -Iseconds)\"" >> ./$(VERSIONGO)
		echo "const GOWEB_BUILD_COMMIT=\"$(shell git log -1 --pretty='%H')\"" >> ./$(VERSIONGO)
		go build -o ../static/page.wasm
