# signalwire-golang
Go SDK

# linter

## installing linter:
```
go get -u github.com/mgechev/revive
go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
```

## running linter inside each package (git repo subfolder):
```
revive -config revive.toml -formatter default
golangci-lint run
```

## running mockgen (generate mock interfaces for unit-testing)
```
cd signalwire-golang/signalwire
mockgen -source=blade.go -destination=blade_mock.go -package=signalwire
mockgen -source=call.go -destination=call_mock.go -package=signalwire
mockgen -source=relay_calling.go -destination=relay_calling_mock.go -package=signalwire
mockgen -source=event_dispatcher.go -destination=event_dispatcher_mock.go -package=signalwire
mockgen -source=client.go -destination=client_mock.go -package=signalwire
```

