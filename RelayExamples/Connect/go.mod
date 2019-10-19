module github.com/signalwire/signalwire-golang/RelayExamples/Outbound

go 1.12

replace github.com/signalwire/signalwire-golang/signalwire => ../../signalwire

require (
	github.com/signalwire/signalwire-golang/signalwire v0.0.0-20190927093955-a506b8d3178f
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/sys v0.0.0-20190927073244-c990c680b611 // indirect
	gopkg.in/yaml.v2 v2.2.3 // indirect
)
