module gitlab.myinterest.top/security/agent/test_collector

go 1.25

replace business_plugins/lib => ../business_plugins/lib

require (
	business_plugins/lib v0.0.0
	gitlab.myinterest.top/security/agent v0.0.0
	go.uber.org/zap v1.27.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace gitlab.myinterest.top/security/agent => ../
