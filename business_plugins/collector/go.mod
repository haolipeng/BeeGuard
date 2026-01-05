module gitlab.myinterest.top/security/agent/business_plugins/collector

go 1.25

replace business_plugins/lib => ../lib

require (
	business_plugins/lib v0.0.0-00010101000000-000000000000
	github.com/go-logr/zapr v1.3.0
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/tklauser/go-sysconf v0.3.16
	go.uber.org/zap v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)
