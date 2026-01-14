module gitlab.myinterest.top/security/agent/business_plugins/collector

go 1.25

replace business_plugins/lib => ../lib

require (
	business_plugins/lib v0.0.0-00010101000000-000000000000
	github.com/GehirnInc/crypt v0.0.0-20230320061759-8cc1b52080c5
	github.com/deckarep/golang-set/v2 v2.8.0
	github.com/docker/docker v20.10.21+incompatible
	github.com/go-logr/zapr v1.3.0
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/karrick/godirwalk v1.17.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/tklauser/go-sysconf v0.3.16
	go.uber.org/zap v1.27.1
	golang.org/x/sys v0.38.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/Microsoft/go-winio v0.4.21 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/tklauser/numcpus v0.11.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
