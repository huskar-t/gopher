package connector

import (
	"github.com/huskar-t/gopher/common/define/tsdb"
	"time"
)

type Data struct {
	Head []string        `json:"head"`
	Data [][]interface{} `json:"data"`
}
type Field struct {
	Key   string
	Value interface{}
	Type  tsdb.PointType
	TS    time.Time
}
type TDEngineConnector interface {
	Exec(sql string) (*Data, error)
	Save(db, table string, fields []*Field) error
}
