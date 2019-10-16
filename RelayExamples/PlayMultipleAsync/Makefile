GO_BIN ?= go
OUT_BIN = PlayMultipleAsync

export PATH := $(PATH):/usr/local/go/bin

all: clean build

build:
	$(GO_BIN) mod tidy
	$(GO_BIN) build -ldflags="-s -w" -o $(OUT_BIN) -v

update:
	$(GO_BIN) get -u
	$(GO_BIN) mod tidy
	$(GO_BIN) mod vendor

clean:
	$(GO_BIN) clean
	rm -f $(OUT_BIN)

test:
	$(GO_BIN) test -failfast

lint:
	golangci-lint run
	revive -config revive.toml
