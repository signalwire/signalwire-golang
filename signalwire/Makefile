GO_BIN ?= go

export PATH := $(PATH):/usr/local/go/bin

all: clean

update:
	$(GO_BIN) get -u
	$(GO_BIN) mod tidy
	$(GO_BIN) mod vendor

clean:
	$(GO_BIN) clean

test:
	$(GO_BIN) test -failfast

lint:
	golangci-lint run
	revive -config revive.toml
