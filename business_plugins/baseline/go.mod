module baseline

go 1.25

require business_plugins/lib v0.0.0

require github.com/gogo/protobuf v1.3.2 // indirect

replace business_plugins/lib => ../lib
