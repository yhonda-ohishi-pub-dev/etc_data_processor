module github.com/yhonda-ohishi/etc_data_processor

go 1.25.1

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
	github.com/yhonda-ohishi/db_service v0.0.0-20251018073811-e72f955d8ce8
	golang.org/x/text v0.29.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250922171735-9219d122eba9
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
)

replace github.com/yhonda-ohishi/db_service => ../db_service
