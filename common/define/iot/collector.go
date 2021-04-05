package iot

//采集器  dtu 串口 tcp
type Collector interface {
	Name() string
	Type() string
	Address() string
	Start() error
	Stop() error
	Connected() bool
	IsRunning() bool
}
