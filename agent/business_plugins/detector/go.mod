module github.com/haolipeng/BeeGuard/agent/business_plugins/detector

go 1.25

replace business_plugins/lib => ../lib

replace shared/datatype => ../../../shared/datatype

require (
	shared/datatype v0.0.0
	business_plugins/lib v0.0.0-00010101000000-000000000000
	github.com/nxadm/tail v1.4.11
	go.uber.org/zap v1.27.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)
