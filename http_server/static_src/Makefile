export GONOPROXY="github.com/AnimusPEXUS/*"
export GOOS=js
export GOARCH=wasm

all: get build

get:
		go get -u -v "./..."
		go mod tidy

build:
		go build -v -x -o ../static/page.wasm
