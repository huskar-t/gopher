package iot

//控制器

type Device struct {
	DeviceID        string
	EdgeID          string
	Points          []*Point
	GatewayID       string `json:"gatewayID"`       // 网关序列号
	DataChannelName string `json:"dataChannelName"` // 关联采集通道名称
	SlaveID         int    `json:"slaveID"`         // 默认的设备从站地址
	IdleTimeout     int    `json:"idleTimeout"`     // 设备空闲时长，用于判断设备离线
}
type Point struct {
	Name string
	Type string
	Value interface{}
}

type Controller interface {
	DeviceCount() int
	AddDevice(device *Device)
	BatchRemoteControl(edgeID string, deviceID string, fields map[string]string) (int, error)
	CallData(request []*Device) (points []*Point, err error)
}
