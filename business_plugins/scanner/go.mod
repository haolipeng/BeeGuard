module scanner

go 1.25

require (
	business_plugins/lib v0.0.0
	github.com/h2non/filetype v1.1.3
	go.uber.org/zap v1.21.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
)

replace business_plugins/lib => ../lib
