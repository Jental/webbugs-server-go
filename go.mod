module webbugs-server

go 1.16

require (
	github.com/ambelovsky/gosf-socketio v0.0.0-20201109193639-add9d32f8b19
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jental/webbugs-common-go v0.0.0-20210705161509-6659b3e53c89 // indirect
)

replace github.com/jental/webbugs-common-go => ../webbugs-common-go