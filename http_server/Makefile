export GONOPROXY="github.com/AnimusPEXUS/*"

all: get build

get:
		$(MAKE) -C static_src get
		go get -u -v "./..."
		go mod tidy

build:
		$(MAKE) -C static_src build
		go build -v -x

run:
		./http_server
