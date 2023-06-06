export GOOS=js
export GOARCH=wasm

all:
		-rm page.wasm
		go get -u -v "./..."
		go mod tidy
		go build -v -o ../static/page.wasm
