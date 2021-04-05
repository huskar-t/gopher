package connector

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/huskar-t/gopher/common/define/tsdb"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type HTTPAuthType string

const (
	BasicAuthType HTTPAuthType = "Basic"
	TaosdAuthType HTTPAuthType = "Taosd"
)

type Scheme string

const (
	HTTPScheme  = "http"
	HTTPSScheme = "https"
)

type TDEngineHTTPResp struct {
	Status     string          `json:"status"`
	Head       []string        `json:"head"` //从 2.0.17 版本开始，建议不要依赖 head 返回值来判断数据列类型，而推荐使用 column_meta。在未来版本中，有可能会从返回值中去掉 head 这一项
	Data       [][]interface{} `json:"data"`
	ColumnMeta [][]interface{} `json:"column_meta"` //从 2.0.17 版本开始，返回值中增加这一项来说明 data 里每一列的数据类型。具体每个列会用三个值来说明，分别为：列名、列类型、类型长度
	Rows       int             `json:"rows"`
	Code       int             `json:"code"`
	Desc       string          `json:"desc"`
}
type HTTPConnector struct {
	scheme     Scheme
	addr       string
	port       string
	authType   HTTPAuthType
	username   string
	password   string
	token      string
	url        url.URL
	httpClient *http.Client
}

func (h *HTTPConnector) Save(db, table string, fields []*Field) error {
	keyMap := map[string]int{}
	TSMap := map[string][]*Field{}
	for _, field := range fields {
		ts := fmt.Sprintf("'%s'",field.TS.In(time.Local).Format("2006-01-02 15:04:05"))
		TSMap[ts] = append(TSMap[ts], field)
		keyMap[field.Key] = 0
	}
	keyList := make([]string, len(keyMap))
	i := 0
	for key := range keyMap {
		keyMap[key] = i + 1
		keyList[i] = key
		i += 1
	}
	values :=[]string{}
	for ts, fieldList := range TSMap {
		value := make([]string,len(keyList)+1)
		value[0] = ts
		for _, field := range fieldList {
			valueString := ""
			switch field.Type {
			case tsdb.PointTypeString:
				valueString = fmt.Sprintf("'%s'",field.Value)
			case tsdb.PointTypeFloat,tsdb.PointTypeBool:
				valueString = fmt.Sprintf("%v",field.Value)
			case tsdb.PointTypeInt, tsdb.PointTypeByte:
				valueString = fmt.Sprintf("%d",field.Value)
			}
			value[keyMap[field.Key]] = valueString
		}
		values = append(values, strings.Join(value,", "))
	}
	sql := fmt.Sprintf("INSERT INTO %s.%s (ts, %s) VALUES (%s)", db, table, strings.Join(keyList, ", "),strings.Join(values,"),("))
	_,err := h.Exec(sql)
	return err
}

func NewHTTPConnector(scheme Scheme, addr string, port string, authType HTTPAuthType, username string, password string) (*HTTPConnector, error) {
	connector := &HTTPConnector{scheme: scheme, addr: addr, port: port, authType: authType, username: username, password: password}
	connector.url = url.URL{
		Scheme: string(scheme),
		Host:   addr + ":" + port,
	}
	connector.httpClient = &http.Client{}
	switch authType {
	case BasicAuthType:
		connector.token = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	case TaosdAuthType:
		loginUrl := path.Join(connector.url.String(), "/rest/login", username, password)
		resp, err := connector.httpClient.Get(loginUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("get taos token error statusCode: %d,body: %s", resp.StatusCode, string(body))
		}
		var respData TDEngineHTTPResp
		err = json.Unmarshal(body, &respData)
		if err != nil {
			return nil, err
		}
		if respData.Status == "succ" {
			connector.token = respData.Desc
		} else {
			return nil, fmt.Errorf("get taos token error statusCode: %d,body: %s", resp.StatusCode, body)
		}
	}
	return connector, nil
}

func (h *HTTPConnector) Exec(sql string) (*Data, error) {
	data, err := h.Query(sql)
	if err != nil {
		return nil, err
	}
	return &Data{
		Head: data.Head,
		Data: data.Data,
	}, nil
}

func (h *HTTPConnector) Query(sql string) (*TDEngineHTTPResp, error) {
	sqlPath := "/rest/sql"
	code, content, err := h.Post(sqlPath, []byte(sql))
	if err != nil {
		if content != nil {
			fmt.Println(string(content))
		}
		return nil, err
	}
	if code != 200 {
		if content != nil {
			return nil, errors.New(string(content))
		}
		return nil, errors.New("return code:" + strconv.Itoa(code))
	}
	var data TDEngineHTTPResp
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	if data.Status != "succ" {
		if data.Desc != "" {
			return &data, fmt.Errorf("code: %d, desc: %s", data.Code, data.Desc)
		}
		return &data, fmt.Errorf("query: %s error,response body: %#v", sql, data)
	}
	return &data, nil
}

func (h *HTTPConnector) Post(subPath string, content []byte) (int, []byte, error) {
	contentReader := bytes.NewReader(content)

	u := h.url
	u.Path = path.Join(u.Path, subPath)
	request, _ := http.NewRequest("POST", u.String(), contentReader)
	if h.token != "" {
		request.Header.Set("Authorization", fmt.Sprintf("%s %s", h.authType, h.token))
	}

	resp, err := h.httpClient.Do(request)
	if err != nil {
		return 0, nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, body, err
	}
	return resp.StatusCode, body, nil
}
