package tsdb

import (
	"time"
)

type PointType string

const (
	PointTypeInt    PointType = "int"
	PointTypeFloat  PointType = "float"
	PointTypeString PointType = "string"
	PointTypeBool   PointType = "bool"
	PointTypeByte   PointType = "byte"
)
var typeMap = map[string]PointType{
	"int":PointTypeInt,
	"float":PointTypeFloat,
	"string":PointTypeString,
	"bool":PointTypeBool,
	"byte":PointTypeByte,
}

func String2PointType(str string) PointType {
	t,ok := typeMap[str]
	if ok {
		return t
	}else {
		return ""
	}
}
type PointInfo struct {
	Name   string    `json:"name"`
	Type   PointType `json:"type"`
	MaxLen int       `json:"maxLen"`
}
type PointDate struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TS    time.Time   `json:"ts"`
}
type DeviceData struct {
	EdgeID   string                 `json:"edgeID"`
	DeviceID string                 `json:"deviceID"`
	TS       time.Time              `json:"ts"`
	Points   map[string]interface{} `json:"points"`
}
type Fill string

const (
	// 线性填充
	FillLine Fill = "line"
	// 前值填充
	FillPrevious Fill = "previous"
	// 空值填充
	FillNull Fill = "null"
	// 不填充
	FillNone Fill = "none"
)

// Aggregation 聚合类型
type Aggregation string

const (
	// 第一条
	AggregationFirst Aggregation = "first"
	// 最后一条
	AggregationLast Aggregation = "last"
	// 中位数
	AggregationMedian Aggregation = "median"
	// 平均值
	AggregationMean Aggregation = "mean"
	// 最大值
	AggregationMax Aggregation = "max"
	// 最小值
	AggregationMin Aggregation = "min"
	// 数量
	AggregationCount Aggregation = "count"
	// 和值
	AggregationSum Aggregation = "sum"
)

//时序库接口定义 Interface definition of TSDB
type TSDB interface {
	CreateDeviceTable(edgeID string, deviceID string, points []*PointInfo, retention string) error
	// SaveTSData 保存时序数据
	SaveTSData(edgeID string,deviceID string, data []*PointDate) error
	// FindLastValueBeforeTime 查询测点最新快照
	FindLastValueBeforeTime(edgeID, deviceID, pointID string, end time.Time) (*PointDate, error)
	// DeleteDeviceData 删除设备数据
	DeleteDeviceData(edgeID, deviceID string) error
	// FindPointData 查询时间范围内的测点历史数据
	QueryPointData(edgeID, deviceID, pointID string, offset, limit int, start, end time.Time) ([]*PointDate, error)
	// QueryDeviceData 查询时间范围内的设备数据
	QueryDeviceData(edgeID, deviceID string, start, end time.Time) ([]*DeviceData, error)
	// QueryPointDataGroupByTime 按时间分组查询测点数据
	QueryPointDataGroupByTime(edgeID, deviceID, pointID string, start, end time.Time, offset, limit int, timeInterval string, fill Fill, aggregations []Aggregation) ([]map[string]interface{}, error)
	// QueryDeviceSnapshotData 查询设备快照数据
	QueryDeviceSnapshotData(edgeID, deviceID, timeInterval string, fill Fill, pointIDArray []string, start, end time.Time) ([]map[string]interface{}, error)
	// QueryDeviceStatisticData 查询设备统计数据
	QueryDeviceStatisticData(edgeID, deviceID, timeInterval, fill Fill, pointIDAggregations map[string][]Aggregation, start, end time.Time) ([]map[string]interface{}, error)
}
