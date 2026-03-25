module github.com/haolipeng/BeeGuard/agent/test_detector

go 1.25

replace business_plugins/lib => ../../../business_plugins/lib

require (
	business_plugins/lib v0.0.0
	github.com/haolipeng/BeeGuard/agent v0.0.0
	go.uber.org/zap v1.27.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/haolipeng/BeeGuard/agent => ../../../
