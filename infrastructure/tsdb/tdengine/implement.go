package tdengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/huskar-t/gopher/common/define/cache"
	"github.com/huskar-t/gopher/common/define/tsdb"
	"github.com/huskar-t/gopher/infrastructure/tsdb/tdengine/connector"
	"strings"
	"time"
)

type TDEngine struct {
	connector connector.TDEngineConnector
	typeCache cache.Cache
}

func (t *TDEngine) CreateDeviceTable(edgeID string, deviceID string, points []*tsdb.PointInfo, retention string) error {
	if retention != "" { // 设置默认存储策略
		keep := strings.TrimSuffix(retention, "d")
		_, err := t.connector.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s KEEP '%v'`, edgeID, keep))
		if err != nil {
			return err
		}
	} else {
		_, err := t.connector.Exec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, edgeID))
		if err != nil {
			return err
		}
	}
	createTable, err := t.buildCreateTableSQL(deviceID, points)
	if err != nil {
		return err
	}
	_, err = t.connector.Exec(createTable)
	if err != nil {
		return err
	}
	return nil
}

func (t *TDEngine) SaveTSData(edgeID string, deviceID string, data []*tsdb.PointDate) error {
	var fields []*connector.Field
	for _, item := range data {
		pt, err := t.getPointType(edgeID, deviceID, item.Key)
		if err != nil {
			return err
		}
		fields = append(fields, &connector.Field{
			Key:   item.Key,
			Value: item.Value,
			Type:  pt,
			TS:    item.TS,
		})
	}
	return t.connector.Save(edgeID, deviceID, fields)
}

func (t *TDEngine) FindLastValueBeforeTime(edgeID, deviceID, pointID string, end time.Time) (*tsdb.PointDate, error) {
	sql := fmt.Sprintf("select ts , last(%s) from %s.%s where ts <='%s'", pointID, edgeID, deviceID, t.formatTime(end))
	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 || len(resp.Data[0]) == 0 {
		return nil, errors.New(fmt.Sprintf("exec error %#v", resp.Data))
	}
	value := resp.Data[0]
	return &tsdb.PointDate{
		Key:   pointID,
		Value: value[1],
		TS:    t.parseTime(value[0].(string)),
	}, nil
}

func (t *TDEngine) DeleteDeviceData(edgeID, deviceID string) error {
	_, err := t.connector.Exec(fmt.Sprintf("DROP TABLE IF EXISTS '%s.%v'", edgeID, deviceID))
	if err != nil {
		return err
	}
	return nil
}

func (t *TDEngine) QueryPointData(edgeID, deviceID, pointID string, offset, limit int, start, end time.Time) ([]*tsdb.PointDate, error) {
	var buffer bytes.Buffer
	_, _ = fmt.Fprintf(&buffer, `Select %s,ts FROM %s.%s WHERE ts > '%s' AND ts <= '%s'`, pointID, edgeID, deviceID, t.formatTime(start), t.formatTime(end))
	if limit > 0 {
		_, _ = fmt.Fprintf(&buffer, " limit %d offset %d", limit, offset)
	}

	sql := buffer.String()
	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}

	dataList := make([]*tsdb.PointDate, len(resp.Data))
	for i, data := range resp.Data {
		dataList[i] = &tsdb.PointDate{
			Key:   pointID,
			Value: data[0],
			TS:    t.parseTime(data[1].(string)),
		}
	}
	return dataList, nil
}

func (t *TDEngine) QueryDeviceData(edgeID, deviceID string, start, end time.Time) ([]*tsdb.DeviceData, error) {
	sql := fmt.Sprintf(`select * FROM '%s.%s'
	WHERE ts > '%s' AND ts <= '%s'
	ORDER BY ts`, edgeID, deviceID, t.formatTime(start), t.formatTime(end))
	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}

	dataList := make([]*tsdb.DeviceData, len(resp.Data))
	for index, data := range resp.Data {
		item := &tsdb.DeviceData{
			Points: map[string]interface{}{},
		}
		for i, head := range resp.Head {
			if head == "ts" {
				item.TS = t.parseTime(data[i].(string))
			} else {
				item.Points[head] = data[i]
			}
		}
		dataList[index] = item
	}
	return dataList, nil
}

func (t *TDEngine) QueryPointDataGroupByTime(
	edgeID, deviceID, pointID string,
	start, end time.Time,
	offset, limit int,
	timeInterval string,
	fill tsdb.Fill,
	aggregations []tsdb.Aggregation,
) ([]map[string]interface{}, error) {
	var stringBuilder []string
	for _, aggregation := range aggregations {
		stringBuilder = append(stringBuilder, t.aggregationAs(aggregation, pointID))
	}

	var buffer bytes.Buffer
	fmt.Fprintf(&buffer, `select %s from %s.%s
	where ts > '%s' and ts <= '%s'
	INTERVAL(%s)
	fill(%s)`,
		strings.Join(stringBuilder, ", "), edgeID, deviceID, t.formatTime(start), t.formatTime(end), timeInterval, t.changeFill(fill))
	if limit > 0 {
		fmt.Fprintf(&buffer, " limit %d offset %d", limit, offset)
	}

	sql := buffer.String()

	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}
	var records []map[string]interface{}
	for _, data := range resp.Data {
		item := map[string]interface{}{}
		for i, head := range resp.Head {
			if head == "ts" {
				item["time"] = t.parseTime(data[i].(string))
			} else {
				item[head] = data[i]
			}
		}
		records = append(records, item)
	}

	return records, nil
}

func (t *TDEngine) QueryDeviceSnapshotData(edgeID, deviceID, timeInterval string, fill tsdb.Fill, pointIDArray []string, start, end time.Time) ([]map[string]interface{}, error) {
	var sql string
	if timeInterval != "" {
		var pointIDBuilder []string
		for _, pointID := range pointIDArray {
			pointIDBuilder = append(pointIDBuilder, fmt.Sprintf(`FIRST(%s) as %s`, pointID, pointID))
		}

		sql = fmt.Sprintf(`select %s from %s.%s
	where ts > '%s' and ts <= '%s'
	INTERVAL(%s)
	fill(%s)`,
			strings.Join(pointIDBuilder, ", "),
			edgeID, deviceID, t.formatTime(start), t.formatTime(end), timeInterval, t.changeFill(fill))
	} else {
		sql = fmt.Sprintf(`select %s from %s.%s where ts > '%s' and ts <= '%s'`,
			strings.Join(pointIDArray, ", "), edgeID, deviceID, t.formatTime(start), t.formatTime(end))
	}
	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}
	var records []map[string]interface{}
	for _, data := range resp.Data {
		item := map[string]interface{}{}
		for i, head := range resp.Head {
			if head == "ts" {
				item["time"] = t.parseTime(data[i].(string))
			} else {
				item[head] = data[i]
			}
		}
		records = append(records, item)
	}
	return records, nil
}

func (t *TDEngine) QueryDeviceStatisticData(edgeID, deviceID, timeInterval, fill tsdb.Fill, pointIDAggregations map[string][]tsdb.Aggregation, start, end time.Time) ([]map[string]interface{}, error) {
	var stringBuilder []string
	for pointID, items := range pointIDAggregations {
		for _, item := range items {
			stringBuilder = append(stringBuilder, t.aggregationAs(item, pointID))
		}
	}
	sql := fmt.Sprintf(`select %s from %s.%s
	where ts > '%s' and ts <= '%s'
	INTERVAL(%s)
	fill(%s)`,
		strings.Join(stringBuilder, ", "), edgeID, deviceID, t.formatTime(start), t.formatTime(end), timeInterval, t.changeFill(fill))
	resp, err := t.connector.Exec(sql)
	if err != nil {
		return nil, err
	}
	var records []map[string]interface{}
	for _, data := range resp.Data {
		item := map[string]interface{}{}
		for i, head := range resp.Head {
			if head == "ts" {
				item["time"] = t.parseTime(data[i].(string))
			} else {
				item[head] = data[i]
			}
		}
		records = append(records, item)
	}
	return records, nil
}

func (t *TDEngine) buildCreateTableSQL(deviceID string, points []*tsdb.PointInfo) (string, error) {
	columnTypes := make([]string, len(points)+1)
	columnTypes[0] = `ts TIMESTAMP`
	for i, point := range points {
		if point.Name == "ts" {
			return "", errors.New("ts is a reserved field")
		}
		var columnType string
		switch point.Type {
		case tsdb.PointTypeInt:
			columnType = fmt.Sprintf("%s INT", point.Name)
		case tsdb.PointTypeFloat:
			columnType = fmt.Sprintf("%s FLOAT", point.Name)
		case tsdb.PointTypeString:
			if point.MaxLen == 0 {
				point.MaxLen = 128
			}
			columnType = fmt.Sprintf("%s(%d) NCHAR", point.Name, point.MaxLen)
		case tsdb.PointTypeBool:
			columnType = fmt.Sprintf("%s BOOL", point.Name)
		case tsdb.PointTypeByte:
			columnType = fmt.Sprintf("%s SMALLINT", point.Name)
		default:
			return "", errors.New("unknow type:" + string(point.Type))
		}
		columnTypes[i+1] = columnType
	}
	sql := fmt.Sprintf("%s %s (%s)", `CREATE TABLE IF NOT EXISTS`, deviceID, strings.Join(columnTypes, ","))
	return sql, nil
}

func (t *TDEngine) formatTime(in time.Time) string {
	return in.In(time.Local).Format("2006-01-02 15:04:05.000")
}
func (t *TDEngine) parseTime(ts string) time.Time {
	out, _ := time.ParseInLocation("2006-01-02 15:04:05.000", ts, time.Local)
	return out
}
func (t *TDEngine) aggregationAs(aggregation tsdb.Aggregation, pointID string) string {
	//"first", "last", "median", "mean", "max", "mean", "count", "sum"
	//精度
	switch aggregation {
	case tsdb.AggregationMedian:
		return fmt.Sprintf("percentile(%s,50) as median", pointID)
	case tsdb.AggregationMean:
		return fmt.Sprintf("avg(%s) as mean", pointID)
	case tsdb.AggregationFirst,
		tsdb.AggregationLast,
		tsdb.AggregationMin,
		tsdb.AggregationMax,
		tsdb.AggregationCount,
		tsdb.AggregationSum:
		return fmt.Sprintf("%s(%s) as %s", aggregation, pointID, aggregation)
	default:
		return ""
	}
}
func (t *TDEngine) changeFill(fill tsdb.Fill) string {
	switch fill {
	case tsdb.FillPrevious:
		return "PREV"
	case tsdb.FillLine:
		return "LINEAR"
	case tsdb.FillNone:
		return "NULL"
	case tsdb.FillNull:
		return "NONE"
	default:
		return "NONE"
	}
}
func (t *TDEngine) getPointType(edgeID, deviceID, pointID string) (tsdb.PointType, error) {
	var ptStr string
	exist,err := t.typeCache.Get(context.TODO(), fmt.Sprintf("%s:%s:%s", edgeID, deviceID, pointID), &ptStr)
	if err != nil {
		return "", err
	}
	if !exist{
		return "", errors.New("")
	}
	if ptStr == "" {
		return "", errors.New("get type error")
	}
	pt := tsdb.String2PointType(ptStr)
	if pt == "" {
		return "", errors.New("change point type error")
	}
	return pt, nil
}
