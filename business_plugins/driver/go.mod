module driver

go 1.25

require (
	business_plugins/lib v0.0.0
	github.com/cilium/ebpf v0.12.0
	go.uber.org/zap v1.21.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // indirect
	golang.org/x/sys v0.38.0 // indirect
)

replace business_plugins/lib => ../lib
