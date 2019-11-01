GO_BIN ?= go

export PATH := $(PATH):/usr/local/go/bin

linter-install:
	$(GO_BIN) get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GO_BIN) get -u github.com/mgechev/revive

test:
	make test -C signalwire

lint:
	make lint -C signalwire
	make lint -C RelayTests/CallComplex
	make lint -C RelayTests/CallOutbound
	make lint -C RelayTests/CallInbound
	make lint -C RelayTests/CallRecord
	make lint -C RelayExamples/Outbound
	make lint -C RelayExamples/Inbound
	make lint -C RelayExamples/PlayAsync
	make lint -C RelayExamples/RecordAsync
	make lint -C RelayExamples/RecordMultipleAsync
	make lint -C RelayExamples/PlayMultipleAsync
	make lint -C RelayExamples/Detect
	make lint -C RelayExamples/ReceiveFax
	make lint -C RelayExamples/SendFax
	make lint -C RelayExamples/RecordBlocking
	make lint -C RelayExamples/ReceiveFaxBlocking
	make lint -C RelayExamples/Connect
	make lint -C RelayExamples/Tap
	make lint -C RelayExamples/SendDigits
	make lint -C RelayExamples/MessageSend
	make lint -C RelayExamples/DeliverTask

update:
	make update -C signalwire
	make update -C RelayTests/CallComplex
	make update -C RelayTests/CallOutbound
	make update -C RelayTests/CallInbound
	make update -C RelayTests/CallRecord
	make update -C RelayExamples/Outbound
	make update -C RelayExamples/Inbound
	make update -C RelayExamples/PlayAsync
	make update -C RelayExamples/RecordAsync
	make update -C RelayExamples/RecordMultipleAsync
	make update -C RelayExamples/PlayMultipleAsync
	make update -C RelayExamples/Detect
	make update -C RelayExamples/ReceiveFax
	make update -C RelayExamples/SendFax
	make update -C RelayExamples/RecordBlocking
	make update -C RelayExamples/ReceiveFaxBlocking
	make update -C RelayExamples/Connect
	make update -C RelayExamples/Tap
	make update -C RelayExamples/SendDigits
	make update -C RelayExamples/MessageSend
	make update -C RelayExamples/DeliverTask

