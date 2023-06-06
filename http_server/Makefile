export GONOPROXY="github.com/AnimusPEXUS/*"

all:
		$(MAKE) -C static_src
		go get -u -v "./..."
		go mod tidy
		go build
