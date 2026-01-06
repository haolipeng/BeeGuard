module test

go 1.25

require (
	business_plugins/lib v0.0.0
	gitlab.myinterest.top/security/agent v0.0.0
	go.uber.org/zap v1.21.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)

replace gitlab.myinterest.top/security/agent => ..

replace business_plugins/lib => ../business_plugins/lib
