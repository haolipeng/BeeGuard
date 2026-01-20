package container

type State int32

const (
	CREATED State = 0 //已创建
	RUNNING State = 1 //运行中
	EXITED  State = 2 //已退出
	UNKNOWN State = 3 //未知
)

var StateName = map[int32]string{
	0: "created",
	1: "running",
	2: "exited",
	3: "unknown",
}

var StateValue = map[string]int32{
	"created": 0,
	"running": 1,
	"exited":  2,
	"unknown": 3,
}
