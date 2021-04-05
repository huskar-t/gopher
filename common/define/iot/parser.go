package iot

//解析器
type Parser interface {
	Name() string
	Protocol() string
	Parse(msg []byte)
}