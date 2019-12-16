module github.com/signalwire/signalwire-golang/RelayExamples/Tap

go 1.12

replace github.com/signalwire/signalwire-golang/signalwire => ../../signalwire

require (
	github.com/gorilla/websocket v1.4.1
	github.com/pion/rtp v1.1.4
	github.com/signalwire/signalwire-golang/signalwire v0.0.0-20190927093955-a506b8d3178f
	github.com/sirupsen/logrus v1.4.2
	github.com/zaf/g711 v0.0.0-20190814101024-76a4a538f52b
	golang.org/x/tools v0.0.0-20191216173652-a0e659d51361 // indirect
	gopkg.in/hraban/opus.v2 v2.0.0-20191117073431-57179dff69a6
)
